/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/
package main

import (
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
)

func MainLoop(app *gopi.AppInstance, done chan<- struct{}) error {
	// Create a new remote keymap
	device := remotes.NewRemote(remotes.CODEC_NEC32, 0)

	// Set keys
	if err := device.SetKey(remotes.KEYCODE_POWER_TOGGLE, 0x40, ""); err != nil {
		return err
	}
	if err := device.SetKey(remotes.KEYCODE_VOLUME_UP, 0x80, ""); err != nil {
		return err
	}
	if err := device.SetKey(remotes.KEYCODE_VOLUME_DOWN, 0x00, ""); err != nil {
		return err
	}
	if err := device.SetKey(remotes.KEYCODE_PLAY, 0x10, ""); err != nil {
		return err
	}

	// Save file
	if err := device.Save(); err != nil {
		return err
	}

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Configuration
	config := gopi.NewAppConfig()
	config.AppFlags.FlagString("device", "", "Name of device to learn")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, MainLoop))
}
