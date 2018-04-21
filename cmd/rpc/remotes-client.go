/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"

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

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	client := app.ModuleInstance("rpc/client:grpc").(gopi.RPCClientConn)
	start = make(chan pb.RemotesClient)

	if services, err := client.Connect(); err != nil {
		done <- gopi.DONE
		return err
	} else if has_service := HasService(services, "MiHome"); has_service == false {
		done <- gopi.DONE
		return fmt.Errorf("Invalid remotes gateway address (missing service)")
	} else if obj, err := client.NewService(reflect.ValueOf(pb.NewRemotesClient)); err != nil {
		done <- gopi.DONE
		return err
	} else if service, ok := obj.(pb.RemotesClient); service == nil || ok == false {
		done <- gopi.DONE
		return errors.New("Invalid remotes service")
	} else {
		// Send the service to the receive loop
		// TODO start <- service
	}

	// Wait for signal
	app.Logger.Debug("Waiting for CTRL+C")
	app.WaitForSignal()

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/client:grpc")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
