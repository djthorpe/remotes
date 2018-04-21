/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/protobuf/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	start chan pb.RemotesClient
)

////////////////////////////////////////////////////////////////////////////////
// MAIN

func HasService(services []string, service string) bool {
	if services == nil {
		return false
	}
	for _, v := range services {
		if v == service {
			return true
		}
	}
	return false
}

func RemotesService(app *gopi.AppInstance) (pb.RemotesClient, error) {
	if client := app.ModuleInstance("rpc/clientconn").(gopi.RPCClientConn); client == nil {
		return nil, fmt.Errorf("Missing rpc/clientconn module")
	} else if services, err := client.Connect(); err != nil {
		return nil, err
	} else if has_service := HasService(services, "mutablelogic.Remotes"); has_service == false {
		return nil, fmt.Errorf("Missing Remotes service, gateway has the following services: %v", strings.Join(services, ","))
	} else if obj, err := client.NewService(reflect.ValueOf(pb.NewRemotesClient)); err != nil {
		return nil, err
	} else if service, ok := obj.(pb.RemotesClient); service == nil || ok == false {
		return nil, errors.New("Invalid remotes service")
	} else {
		return service, nil
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Create a start channel used to pass the service on indicating
	// the start of processing
	start = make(chan pb.RemotesClient)

	if service, err := RemotesService(app); err != nil {
		done <- gopi.DONE
		return err
	} else {
		// Initiate the service
		start <- service
	}

	// Wait for signal
	app.Logger.Debug("Waiting for CTRL+C")
	app.WaitForSignal()

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Receive(service pb.RemotesClient, done <-chan struct{}) error {

	// Create the context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Call cancel in the background when done is received
	go func() {
		<-done
		cancel()
	}()

	// Receive a stream of input events from the server
	if stream, err := service.Receive(ctx, &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		for {
			if input_event, err := stream.Recv(); err == io.EOF {
				break
			} else if err != nil {
				return err
			} else {
				fmt.Printf("InputEvent=%v\n", input_event)
			}
		}
	}

	// Success
	return nil
}

func Run(app *gopi.AppInstance, done <-chan struct{}) error {

FOR_LOOP:
	for {
		select {
		case service := <-start:
			return Receive(service, done)
		case <-done:
			break FOR_LOOP
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/clientconn")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, Run))
}
