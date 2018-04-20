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

	// Remotes
	_ "github.com/djthorpe/remotes/appletv"
	_ "github.com/djthorpe/remotes/nec"
	_ "github.com/djthorpe/remotes/panasonic"
	_ "github.com/djthorpe/remotes/sony"
)

var (
	CODECS = []string{"remotes/panasonic", "remotes/appletv", "remotes/nec32", "remotes/sony12", "remotes/sony15"}
	//CODECS             = []string{"remotes/rc5"}
	RCV_TIMEOUT uint32 = 100 // ms
)

////////////////////////////////////////////////////////////////////////////////

func PrintHeader() {
	fmt.Printf("%15s %10s %10s\n", "Codec", "Scancode", "Device")
	fmt.Printf("%15s %10s %10s\n", "---------------", "----------", "----------")
}

func PrintEvent(evt gopi.Event) {
	if event, ok := evt.(*remotes.RemoteEvent); event != nil && ok {
		fmt.Printf("%15s %10s %10s %s\n",
			event.Codec(),
			fmt.Sprintf("0x%X", event.Scancode()),
			fmt.Sprintf("0x%X", event.Device()),
			event.EventType(),
		)
	} else {
		fmt.Println(evt)
	}
}

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

	appletv := app.ModuleInstance("remotes/appletv").(remotes.Codec).Subscribe()
	nec32 := app.ModuleInstance("remotes/nec32").(remotes.Codec).Subscribe()
	panasonic := app.ModuleInstance("remotes/panasonic").(remotes.Codec).Subscribe()
	sony12 := app.ModuleInstance("remotes/sony12").(remotes.Codec).Subscribe()
	sony15 := app.ModuleInstance("remotes/sony15").(remotes.Codec).Subscribe()
	//sony20 := app.ModuleInstance("remotes/sony20").(remotes.Codec).Subscribe()
	//rc5 := app.ModuleInstance("remotes/rc5").(remotes.Codec).Subscribe()

	PrintHeader()

FOR_LOOP:
	for {
		select {
		// case evt := <-sony20:
		// 	PrintEvent(evt)
		case evt := <-appletv:
			PrintEvent(evt)
		case evt := <-nec32:
			PrintEvent(evt)
		case evt := <-sony12:
			PrintEvent(evt)
		case evt := <-sony15:
			PrintEvent(evt)
		case evt := <-panasonic:
			PrintEvent(evt)
		// case evt := <-rc5:
		// 	PrintEvent(evt)
		case <-done:
			break FOR_LOOP
		}
	}

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
