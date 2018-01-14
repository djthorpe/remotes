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
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/nec"
	_ "github.com/djthorpe/remotes/sony"
)

var (
	CODECS             = []string{"remotes/sony12", "remotes/sony15", "remotes/sony20", "remotes/nec32"}
	RCV_TIMEOUT uint32 = 100 // ms
)

////////////////////////////////////////////////////////////////////////////////

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	lirc := app.LIRC

	// Try and set the timeout, ignore if feature is not implemented
	if lirc == nil {
		return errors.New("Missing LIRC module")
	} else if err := lirc.SetRcvTimeout(RCV_TIMEOUT); err != nil && err != gopi.ErrNotImplemented {
		return err
	} else if err != gopi.ErrNotImplemented {
		if lirc.SetRcvTimeoutReports(true); err != nil && err != gopi.ErrNotImplemented {
			return err
		}
	}

	sony := app.ModuleInstance("remotes/sony12").(remotes.Codec)
	if sony == nil {
		return errors.New("Missing Sony Codec")
	}
	edge := sony.Subscribe()

FOR_LOOP:
	for {
		select {
		case evt := <-edge:
			if event, ok := evt.(*remotes.RemoteEvent); event != nil && ok {
				fmt.Printf("%10s %X\n", event.Codec(), event.Scancode())
				if err := sony.Send(event.Scancode(), 2); err != nil {
					app.Logger.Error("%v", err)
				}
			} else {
				fmt.Println(evt)
			}
		case <-done:
			break FOR_LOOP
		}
	}

	// Unsubscribe from edges
	sony.Unsubscribe(edge)

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
