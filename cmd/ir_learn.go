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
	"time"

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
	sony := app.ModuleInstance("remotes/sony").(remotes.Codec)
	if sony == nil {
		return errors.New("Missing Sony Codec")
	}
	edge := sony.Subscribe()

FOR_LOOP:
	for {
		select {
		case evt := <-edge:
			fmt.Println(evt)
		case <-done:
			break FOR_LOOP
		}
	}

	// Unsubscribe from edges
	sony.Unsubscribe(edge)

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Learn(key <-chan gopi.InputEvent) (uint32, bool) {
	select {
	case <-time.After(time.Second * 10):
		// Timeout
		return 0, false
	}
	return 0, true
}

func MainLoop(app *gopi.AppInstance, done chan<- struct{}) error {
	if device, exists := app.AppFlags.GetString("device"); exists == false || device == "" {
		return fmt.Errorf("Missing -device flag")
	} else {
		fmt.Printf("Learning remote \"%v\", Press CTRL+C when done\n", device)

		device_map := remotes.NewDeviceMap(device)
		key_chan := make(chan gopi.InputEvent)

		// Populate device map here
		for k := remotes.KEYCODE_MIN + 1; k < remotes.KEYCODE_MAX; k++ {
			fmt.Println("Press", k, "on remote keypad or wait for timeout")
			Learn(key_chan)
		}

		// Wait for interrupt
		app.WaitForSignal()

		// Close channel
		close(key_chan)

		// Write device map
		device_map.Write(os.Stdout)

		// Finish
		fmt.Printf("\n\nWritten to file\n")
	}

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Configuration
	config := gopi.NewAppConfig(CODECS...)
	config.AppFlags.FlagString("device", "", "Name of device to learn")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, MainLoop, EventLoop))
}
