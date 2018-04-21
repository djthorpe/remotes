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
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	evt "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/protobuf/remotes"
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
	log  gopi.Logger
	done *evt.PubSub
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.remotes>Open{ server=%v }", config.Server)

	this := new(service)
	this.log = log
	this.done = evt.NewPubSub(1)

	// Register service with server
	config.Server.Register(this)

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.remotes>Close{}")

	// Close done
	this.done.Close()
	this.done = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER CODECS

func (this *service) registerCodec(codec remotes.Codec) {
	this.log.Debug2("<grpc.service.remotes>RegisterCodec{ codec=%v }", codec)
}

////////////////////////////////////////////////////////////////////////////////
// RPC SERVICE INTERFACE IMPLEMENTATION

func (this *service) CancelRequests() error {
	this.log.Debug2("<grpc.service.remotes>CancelRequests{}")
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
	timer := time.Tick(5 * time.Second)

	// Get a channel we will use for breaking the loop
	done := this.done.Subscribe()

	// Send until loop is broken
FOR_LOOP:
	for {
		select {
		case <-timer:
			reply := toProtobufInputEvent()
			if err := stream.Send(reply); err != nil {
				this.log.Warn("Receive: error sending: %v: closing request", err)
				break FOR_LOOP
			} else {
				this.log.Info("Sent: %v", reply)
			}
		case <-done:
			break FOR_LOOP
		}
	}

	// Unsubscribe from done signal
	this.done.Unsubscribe(done)

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.remotes{}")
}

////////////////////////////////////////////////////////////////////////////////
// PROTOBUF CONVERSION

func toProtobufInputEvent() *pb.InputEvent {
	return &pb.InputEvent{}
}
