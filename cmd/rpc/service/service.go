/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package service

import (
	"fmt"
	"reflect"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	evt "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/protobuf/remotes"
	ptype "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register service/remotes:grpc
	gopi.RegisterModule(gopi.Module{
		Name:     "service/remotes:grpc",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Service{
				Server: app.ModuleInstance("rpc/server").(gopi.RPCServer),
				App:    app,
			}, app.Logger)
		},
		Run: func(app *gopi.AppInstance, driver gopi.Driver) error {
			// Register codecs with driver. Codecs have OTHER as module type
			// and name starting with "remotes/"
			for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_OTHER) {
				if strings.HasPrefix(module.Name, "remotes/") {
					if codec, ok := app.ModuleInstance(module.Name).(remotes.Codec); ok && codec != nil {
						driver.(*service).registerCodec(codec)
					}
				}
			}
			// Success
			return nil
		},
	})
}

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server gopi.RPCServer
	App    *gopi.AppInstance
}

type service struct {
	log    gopi.Logger
	done   *evt.PubSub
	merger evt.EventMerger
	codecs []remotes.Codec
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.remotes>Open{ server=%v }", config.Server)

	this := new(service)
	this.log = log
	this.done = evt.NewPubSub(1)
	this.codecs = make([]remotes.Codec, 0, 10)
	this.merger = evt.NewEventMerger()

	// Register service with server
	config.Server.Register(this)

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
// REGISTER CODECS

func (this *service) registerCodec(codec remotes.Codec) {
	this.log.Debug2("<grpc.service.remotes>RegisterCodec{ codec=%v }", codec)

	// Append codecs and subscribe
	this.codecs = append(this.codecs, codec)
	this.merger.Add(codec.Subscribe())
}

////////////////////////////////////////////////////////////////////////////////
// RPC SERVICE INTERFACE IMPLEMENTATION

func (this *service) CancelRequests() error {
	this.log.Debug2("<grpc.service.remotes>CancelRequests{}")
	// Cancel any streaming requests
	this.done.Emit(evt.NullEvent)
	return nil
}

func (this *service) GRPCHook() reflect.Value {
	// ensure we conform to pb.RemotesServer
	var _ pb.RemotesServer = &service{}
	return reflect.ValueOf(pb.RegisterRemotesServer)
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
			reply := toProtobufInputEvent(evt.(gopi.InputEvent))
			if err := stream.Send(reply); err != nil {
				this.log.Warn("Receive: error sending: %v: closing request", err)
				break FOR_LOOP
			} else {
				this.log.Info("Sent: %v", reply)
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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.remotes{ codecs=%v }", this.codecs)
}

////////////////////////////////////////////////////////////////////////////////
// PROTOBUF CONVERSION

func toProtobufInputEvent(evt gopi.InputEvent) *pb.InputEvent {
	return &pb.InputEvent{
		Ts:         ptype.DurationProto(evt.Timestamp()),
		DeviceType: pb.InputDeviceType(evt.DeviceType()),
		EventType:  pb.InputEventType(evt.EventType()),
		Scancode:   evt.Scancode(),
		Position:   toProtobufPoint(evt.Position()),
		Relative:   toProtobufPoint(evt.Relative()),
		Slot:       uint32(evt.Slot()),
	}
}

func toProtobufPoint(pt gopi.Point) *pb.Point {
	return &pb.Point{
		X: pt.X,
		Y: pt.Y,
	}
}
