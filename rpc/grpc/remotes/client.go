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
	"io"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi/sys/rpc/grpc"
	remotes "github.com/djthorpe/remotes"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/rpc/protobuf/remotes"
	ptypes "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.RemotesClient
	conn gopi.RPCClientConn
}

type KeyMapInfo struct {
	remotes.KeyMap
	Keys uint
}

type Key struct {
	remotes.KeyMapEntry
}

type InputEvent struct {
	Timestamp  time.Duration
	DeviceType gopi.InputDeviceType
	EventType  gopi.InputEventType
	Keycode    remotes.RemoteCode
}

type Event struct {
	InputEvent
	Key
	KeyMapInfo
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewRemotesClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Client) Conn() gopi.RPCClientConn {
	return this.conn
}

func (this *Client) NewContext() context.Context {
	if this.conn.Timeout() == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), this.conn.Timeout())
		return ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

// Return array of codecs supported
func (this *Client) Codecs() ([]remotes.CodecType, error) {
	if reply, err := this.RemotesClient.Codecs(this.NewContext(), &pb.EmptyRequest{}); err != nil {
		return nil, err
	} else {
		codecs := make([]remotes.CodecType, len(reply.Codec))
		for i := range reply.Codec {
			codecs[i] = remotes.CodecType(reply.Codec[i])
		}
		return codecs, nil
	}
}

// Return array of keymaps learnt
func (this *Client) KeyMaps() ([]*KeyMapInfo, error) {
	if reply, err := this.RemotesClient.KeyMaps(this.NewContext(), &pb.EmptyRequest{}); err != nil {
		return nil, err
	} else {
		keymaps := make([]*KeyMapInfo, len(reply.Keymap))
		for i, keymap := range reply.Keymap {
			keymaps[i] = &KeyMapInfo{
				remotes.KeyMap{
					Name:    keymap.Name,
					Type:    remotes.CodecType(keymap.Codec),
					Device:  keymap.Device,
					Repeats: uint(keymap.Repeats),
				}, uint(keymap.Keys),
			}
		}
		return keymaps, nil
	}
}

// Return keys learnt by keymap name
func (this *Client) Keys(keymap string) ([]*Key, error) {
	var reply *pb.KeysReply
	var err error

	// API call
	if keymap == "" {
		reply, err = this.RemotesClient.LookupKeys(this.NewContext(), &pb.LookupKeysRequest{})
	} else {
		reply, err = this.RemotesClient.Keys(this.NewContext(), &pb.KeysRequest{Keymap: keymap})
	}

	// Check for error
	if err != nil {
		return nil, err
	}

	// Return all keys
	keys := make([]*Key, len(reply.Key))
	for i, key := range reply.Key {
		keys[i] = &Key{
			remotes.KeyMapEntry{
				Name:     key.Name,
				Type:     remotes.CodecType(key.Codec),
				Keycode:  remotes.RemoteCode(key.Keycode),
				Device:   key.Device,
				Scancode: key.Scancode,
				Repeats:  uint(key.Repeats),
			},
		}
	}
	return keys, nil
}

// Receive remote events
func (this *Client) Receive(ctx context.Context, evt chan<- *Event) error {
	// Receive a stream of input events from the server, and transmit them via
	// the event channel
	if stream, err := this.RemotesClient.Receive(ctx, &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		for {
			if msg, err := stream.Recv(); err == io.EOF {
				break
			} else if err != nil {
				return gopiError(err)
			} else {
				ts, _ := ptypes.Duration(msg.Event.Ts)
				evt <- &Event{
					InputEvent{
						Timestamp:  ts,
						DeviceType: gopi.InputDeviceType(msg.Event.DeviceType),
						EventType:  gopi.InputEventType(msg.Event.EventType),
						Keycode:    remotes.RemoteCode(msg.Event.Keycode),
					},
					Key{
						remotes.KeyMapEntry{
							Name:     msg.Key.Name,
							Type:     remotes.CodecType(msg.Key.Codec),
							Device:   msg.Key.Device,
							Scancode: msg.Key.Scancode,
							Keycode:  remotes.RemoteCode(msg.Key.Keycode),
							Repeats:  uint(msg.Key.Repeats),
						},
					},
					KeyMapInfo{
						remotes.KeyMap{
							Name:    msg.Keymap.Name,
							Type:    remotes.CodecType(msg.Keymap.Codec),
							Device:  msg.Keymap.Device,
							Repeats: uint(msg.Keymap.Repeats),
						}, uint(msg.Keymap.Keys),
					},
				}
			}
		}
	}
	return nil
}

// Return keys with one or more search terms and optional
// keymap argument to narrow search to a keymap entries
func (this *Client) LookupKeys(keymap string, terms []string) ([]*Key, error) {
	// There needs to be one or more terms
	if len(terms) == 0 {
		return nil, gopi.ErrBadParameter
	}

	// Construct a hash of keys so we can quickly see
	// if a key is in a keymap
	var keymap_hash map[pb.RemoteCode]bool
	if keymap != "" {
		if keymap_keys, err := this.RemotesClient.Keys(this.NewContext(), &pb.KeysRequest{Keymap: keymap}); err != nil {
			return nil, err
		} else {
			keymap_hash = make(map[pb.RemoteCode]bool, len(keymap_keys.Key))
			for _, key := range keymap_keys.Key {
				keymap_hash[key.Keycode] = true
			}
		}
	}

	// Lookup all keys
	all_keys, err := this.RemotesClient.LookupKeys(this.NewContext(), &pb.LookupKeysRequest{Terms: terms})
	if err != nil {
		return nil, err
	}

	// Return keys matched by device
	keys := make([]*Key, 0, len(all_keys.Key))
	for _, key := range all_keys.Key {
		// If there is a hash, then filter on the hash
		if keymap_hash != nil {
			if _, exists := keymap_hash[key.Keycode]; exists == false {
				continue
			}
		}
		// Append the key
		keys = append(keys, &Key{
			remotes.KeyMapEntry{
				Name:     key.Name,
				Type:     remotes.CodecType(key.Codec),
				Device:   key.Device,
				Scancode: key.Scancode,
				Keycode:  remotes.RemoteCode(key.Keycode),
				Repeats:  uint(key.Repeats),
			},
		})
	}
	return keys, nil
}

/*
// Send a remote scancode
func (this *Client) SendScancode(in *SendScancodeRequest) error {
	return gopi.ErrNotImplemented
}

// Send a remote keycode
func (this *Client) SendKeycode(in *SendKeycodeRequest) error {
	return gopi.ErrNotImplemented
}
*/

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<grpc.client.remotes>{ conn=%v }", this.conn)
}

func (this *Event) String() string {
	return fmt.Sprintf("<grpc.client.remotes.Event>{ input_event=%v key=%v keymap=%v }", this.InputEvent, this.Key, this.KeyMapInfo)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func gopiError(err error) error {
	if err == nil {
		return nil
	} else if grpc.IsErrCanceled(err) {
		return nil
	} else if grpc.IsErrDeadlineExceeded(err) {
		return gopi.ErrDeadlineExceeded
	} else {
		return err
	}
}
