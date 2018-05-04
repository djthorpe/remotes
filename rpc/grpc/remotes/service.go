/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"context"
	"fmt"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/sys/rpc/grpc"
	evt "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/rpc/protobuf/remotes"
	ptype "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server  gopi.RPCServer
	KeyMaps remotes.KeyMaps
}

type service struct {
	log     gopi.Logger
	done    *evt.PubSub
	merger  evt.EventMerger
	codecs  map[remotes.CodecType]remotes.Codec
	keymaps remotes.KeyMaps
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.remotes>Open{ server=%v keymaps=%v }", config.Server, config.KeyMaps)

	this := new(service)
	this.log = log
	this.done = evt.NewPubSub(1)
	this.codecs = make(map[remotes.CodecType]remotes.Codec, 10)
	this.merger = evt.NewEventMerger()
	this.keymaps = config.KeyMaps

	// Register service with GRPC server
	pb.RegisterRemotesServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.remotes>Close{ codecs=%v }", this.codecs)

	// Close done
	this.done.Close()
	this.done = nil

	// Release codecs
	this.merger.Close()
	this.merger = nil
	this.codecs = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER CODECS, LOAD & SAVE KEYMAPS

func (this *service) registerCodec(codec remotes.Codec) {
	this.log.Debug2("<grpc.service.remotes>RegisterCodec{ codec=%v }", codec)

	// Append codec and add to merged channel
	if _, exists := this.codecs[codec.Type()]; exists == false {
		this.codecs[codec.Type()] = codec
		this.merger.Add(codec.Subscribe())
	} else {
		this.log.Warn("RegisterCodec: Ignoring second codec with same type %v", codec.Type())
	}
}

func (this *service) loadKeyMaps() error {
	return this.keymaps.LoadKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		this.log.Info("Loading: %v (%v)", filename, keymap.Name)
	})
}

func (this *service) saveKeyMaps() error {
	return this.keymaps.SaveModifiedKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		this.log.Info("Saving: %v (%v)", filename, keymap.Name)
	})
}

////////////////////////////////////////////////////////////////////////////////
// RPC SERVICE INTERFACE IMPLEMENTATION

func (this *service) CancelRequests() error {
	this.log.Debug2("<grpc.service.remotes>CancelRequests{}")

	// Cancel any streaming requests
	this.done.Emit(evt.NullEvent)

	// Save any modified keymaps
	if err := this.saveKeyMaps(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPC SERVICE REQUESTS

func (this *service) Receive(_ *pb.EmptyRequest, stream pb.Remotes_ReceiveServer) error {
	// Subscribe to the merger channel and the channel used for
	// breaking the loop
	input_events := this.merger.Subscribe()
	cancel_requests := this.done.Subscribe()

	// Send until loop is broken
FOR_LOOP:
	for {
		select {
		case evt := <-input_events:
			if remote_evt, ok := evt.(remotes.RemoteEvent); remote_evt != nil && ok {
				// Look up key presses - there may be more than one for each key press
				entries := this.keymaps.LookupKeyMapEntry(remote_evt.Codec(), remote_evt.Device(), remote_evt.Scancode())
				if len(entries) > 0 {
					// Mapped to one or more keys
					for entry, keymap := range entries {
						if reply := toProtobufReceiveReply(remote_evt, keymap, entry); reply == nil {
							this.log.Warn("Receive: unable to form a reply from input event, ignoring")
						} else if err := stream.Send(reply); err != nil {
							this.log.Warn("Receive: error sending: %v: closing request", err)
							break FOR_LOOP
						}
					}
				} else {
					// Unmapped to any key
					if reply := toProtobufReceiveReply(remote_evt, nil, nil); reply == nil {
						this.log.Warn("Receive: unable to form a reply from input event, ignoring")
					} else if err := stream.Send(reply); err != nil {
						this.log.Warn("Receive: error sending: %v: closing request", err)
						break FOR_LOOP
					}
				}
			} else {
				this.log.Warn("Receive: invalid remotes.RemoteEvent, ignoring: %v", evt)
			}
		case <-cancel_requests:
			break FOR_LOOP
		}
	}

	// Unsubscribe from channels
	this.done.Unsubscribe(cancel_requests)
	this.merger.Unsubscribe(input_events)

	// Return success
	return nil
}

func (this *service) SendScancode(ctx context.Context, in *pb.SendScancodeRequest) (*pb.EmptyReply, error) {
	if codec, exists := this.codecs[remotes.CodecType(in.Codec)]; exists == false {
		this.log.Warn("SendScancode: Bad request: Invalid codec (%v)", remotes.CodecType(in.Codec))
		return nil, gopi.ErrBadParameter
	} else if err := codec.Send(in.Device, in.Scancode, uint(in.Repeats)); err != nil {
		return nil, err
	} else {
		// Success
		return &pb.EmptyReply{}, nil
	}
}

func (this *service) SendKeycode(ctx context.Context, in *pb.SendKeycodeRequest) (*pb.EmptyReply, error) {

	if keymaps := this.keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, in.Keymap); len(keymaps) != 1 {
		// Keymap not found
		this.log.Warn("SendKeycode: Bad request: Invalid keymap (%v)", in.Keymap)
		return nil, gopi.ErrBadParameter
	} else if allkeyentries := this.keymaps.LookupKeyCode(in.Keycode); len(allkeyentries) == 0 {
		// Keycode not found
		this.log.Warn("SendKeycode: Bad request: Invalid keycode (%v)", in.Keycode)
		return nil, gopi.ErrBadParameter
	} else {
		// Lookup entries in the keymap with this keycode
		entries := make([]*remotes.KeyMapEntry, 0, 1)
		for _, keyentry := range allkeyentries {
			entries = append(entries, this.keymaps.GetKeyMapEntry(keymaps[0], remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, keyentry.Keycode, remotes.SCANCODE_UNKNOWN)...)
		}
		if len(entries) == 0 {
			// No key entry found
			this.log.Warn("SendKeycode: Bad request: Keycode not found (%v)", in.Keycode)
			return nil, remotes.ErrNotFound
		}
		if len(entries) > 1 {
			// There are more than one key for that name
			ambigious := ""
			for _, entry := range entries {
				ambigious += fmt.Sprint("'" + entry.Name + "',")
			}
			return nil, fmt.Errorf("Ambiguous: '%v' (It could mean one of %v)", in.Keycode, strings.TrimSuffix(ambigious, ","))
		}
		if codec, exists := this.codecs[entries[0].Type]; exists == false {
			this.log.Warn("SendKeycode: Bad request: Invalid codec (%v)", entries[0].Type)
			return nil, gopi.ErrBadParameter
		} else {
			// If there is a repeats parameter, then use that to override
			repeats := uint(in.Repeats)
			if repeats == 0 {
				repeats = entries[0].Repeats
			}
			// Perform the sending
			if err := codec.Send(entries[0].Device, entries[0].Scancode, repeats); err != nil {
				return nil, err
			}

		}
	}

	// Success
	return &pb.EmptyReply{}, nil
}

func (this *service) Codecs(ctx context.Context, in *pb.EmptyRequest) (*pb.CodecsReply, error) {
	return toProtobufCodecsReply(this.codecs), nil
}

func (this *service) KeyMaps(ctx context.Context, in *pb.EmptyRequest) (*pb.KeyMapsReply, error) {
	return toProtobufKeyMapsReply(this.keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, "")), nil
}

func (this *service) Keys(ctx context.Context, in *pb.KeysRequest) (*pb.KeysReply, error) {
	// Obtain the keymap
	if in.Keymap == "" {
		return nil, gopi.ErrBadParameter
	} else if keymaps := this.keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, in.Keymap); len(keymaps) == 0 {
		return nil, remotes.ErrNotFound
	} else if len(keymaps) > 1 {
		return nil, remotes.ErrAmbiguous
	} else {
		return toProtobufKeysReply(this.keymaps.GetKeyMapEntry(keymaps[0], remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, remotes.KEYCODE_NONE, remotes.SCANCODE_UNKNOWN)), nil
	}
}

func (this *service) LookupKeys(ctx context.Context, in *pb.LookupKeysRequest) (*pb.KeysReply, error) {
	return toProtobufKeysReply(this.keymaps.LookupKeyCode(in.Terms...)), nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.remotes{ codecs=%v keymaps=%v }", this.codecs, this.keymaps)
}

////////////////////////////////////////////////////////////////////////////////
// PROTOBUF CONVERSION

func toProtobufKeysReply(entries []*remotes.KeyMapEntry) *pb.KeysReply {
	reply := &pb.KeysReply{
		Key: make([]*pb.Key, len(entries)),
	}
	for i, entry := range entries {
		reply.Key[i] = &pb.Key{
			Name:     entry.Name,
			Keycode:  fmt.Sprint(entry.Keycode),
			Scancode: entry.Scancode,
			Device:   entry.Device,
			Codec:    pb.CodecType(entry.Type),
			Repeats:  uint32(entry.Repeats),
		}
	}
	return reply
}

func toProtobufCodecsReply(codecs map[remotes.CodecType]remotes.Codec) *pb.CodecsReply {
	reply := &pb.CodecsReply{
		Codec: make([]pb.CodecType, 0, len(codecs)),
	}
	for k := range codecs {
		reply.Codec = append(reply.Codec, pb.CodecType(k))
	}
	return reply
}

func toProtobufKeyMapsReply(keymaps []*remotes.KeyMap) *pb.KeyMapsReply {
	// Create the reply
	reply := &pb.KeyMapsReply{
		Keymap: make([]*pb.KeyMapInfo, 0, len(keymaps)),
	}
	// Populate the reply
	for _, keymap := range keymaps {
		reply.Keymap = append(reply.Keymap, toProtobufKeyMapInfo(keymap))
	}
	// Return the reply
	return reply
}

func toProtobufReceiveReply(evt remotes.RemoteEvent, keymap *remotes.KeyMap, entry *remotes.KeyMapEntry) *pb.ReceiveReply {
	return &pb.ReceiveReply{
		Event:  toProtobufInputEvent(evt, entry),
		Key:    toProtobufKey(evt, entry),
		Keymap: toProtobufKeyMapInfo(keymap),
	}
}

func toProtobufInputEvent(evt gopi.InputEvent, entry *remotes.KeyMapEntry) *pb.InputEvent {
	input_event := &pb.InputEvent{
		Ts:         ptype.DurationProto(evt.Timestamp()),
		DeviceType: pb.InputDeviceType(evt.DeviceType()),
		EventType:  pb.InputEventType(evt.EventType()),
		Device:     evt.Device(),
		Scancode:   evt.Scancode(),
		Position:   toProtobufPoint(evt.Position()),
		Relative:   toProtobufPoint(evt.Relative()),
		Slot:       uint32(evt.Slot()),
	}
	if entry != nil {
		input_event.Keycode = uint32(entry.Keycode)
	}
	return input_event
}

func toProtobufKey(evt remotes.RemoteEvent, entry *remotes.KeyMapEntry) *pb.Key {
	if entry == nil {
		return &pb.Key{
			Codec:    pb.CodecType(evt.Codec()),
			Device:   evt.Device(),
			Scancode: evt.Scancode(),
		}
	} else {
		return &pb.Key{
			Name:     entry.Name,
			Keycode:  fmt.Sprint(entry.Keycode),
			Codec:    pb.CodecType(entry.Type),
			Device:   entry.Device,
			Scancode: entry.Scancode,
			Repeats:  uint32(entry.Repeats),
		}
	}
}

func toProtobufKeyMapInfo(keymap *remotes.KeyMap) *pb.KeyMapInfo {
	if keymap == nil {
		return nil
	} else {
		return &pb.KeyMapInfo{
			Name:    keymap.Name,
			Codec:   pb.CodecType(keymap.Type),
			Device:  keymap.Device,
			Repeats: uint32(keymap.Repeats),
			Keys:    uint32(len(keymap.Map)),
		}
	}
}

func toProtobufPoint(pt gopi.Point) *pb.Point {
	return &pb.Point{
		X: pt.X,
		Y: pt.Y,
	}
}
