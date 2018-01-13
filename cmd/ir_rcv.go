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
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/sony"
)

var (
	CODECS = []string{"remotes/sony"}
)

////////////////////////////////////////////////////////////////////////////////

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	if sony := app.ModuleInstance("remotes/sony").(remotes.Codec); sony == nil {
		return errors.New("Missing Sony Codec")
	}
	/*
	   	edge := app.LIRC.Subscribe()

	   FOR_LOOP:
	   	for {
	   		select {
	   		case evt := <-edge:
	   			fmt.Println("EVENT: ", evt)
	   		case <-done:
	   			break FOR_LOOP
	   		}
	   	}

	   	// Unsubscribe from edges
	   	app.LIRC.Unsubscribe(edge)
	*/

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func MainLoop(app *gopi.AppInstance, done chan<- struct{}) error {

	if app.LIRC == nil {
		return errors.New("Missing LIRC module")
	}

	// Wait for interrupt
	app.WaitForSignal()

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Configuration
	config := gopi.NewAppConfig(CODECS...)

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, MainLoop, EventLoop))
}
