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

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/sony"
)

var (
	CODECS = []string{"remotes/sony"}
)

////////////////////////////////////////////////////////////////////////////////

func EventLoop(app *gopi.AppInstance, done chan struct{}) error {
	if app.LIRC == nil {
		return errors.New("Missing LIRC module")
	}

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
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func MainLoop(app *gopi.AppInstance, done chan struct{}) error {

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

func main_inner() int {
	// Configuration
	config := gopi.NewAppConfig(CODECS...)
	// Create the application
	app, err := gopi.NewAppInstance(config)
	if err != nil {
		if err != gopi.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return -1
		}
		return 0
	}
	defer app.Close()

	// Run the application
	if err := app.Run(MainLoop, EventLoop); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	}
	return 0
}

func main() {
	os.Exit(main_inner())
}
