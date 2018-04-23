/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
)

const (
	DEFAULT_EXT = ".keymap"
)

func Load(name string) (*remotes.Remote, string, error) {
	// Set the name to include extension
	if filepath.Ext(name) != DEFAULT_EXT {
		name = name + DEFAULT_EXT
	}

	// If name is a filename and exists then load it in
	if _, err := os.Stat(name); os.IsNotExist(err) {
		device_name := filepath.Base(name)

		// Strip name of extension
		if strings.HasSuffix(device_name, DEFAULT_EXT) {
			device_name = strings.TrimSuffix(device_name, DEFAULT_EXT)
		}

		// Create a new remote
		remote := remotes.NewRemote(remotes.CODEC_NEC32, 0)
		remote.SetName(device_name)

		// Return values
		return remote, name, nil
	} else if remote, err := remotes.RemoteFromFile(name); err != nil {
		return nil, "", err
	} else {
		return remote, name, nil
	}
}

func MainLoop(app *gopi.AppInstance, done chan<- struct{}) error {
	// If device is empty, then return error
	if device_name, exists := app.AppFlags.GetString("device"); exists == false {
		return fmt.Errorf("Missing -device flag")
	} else if remote, filename, err := Load(device_name); err != nil {
		return err
	} else {
		app.Logger.Info("remote=%v", remote)

		// Set keys
		if err := remote.SetKey(remotes.KEYCODE_POWER_TOGGLE, 0x40, ""); err != nil {
			return err
		}
		if err := remote.SetKey(remotes.KEYCODE_VOLUME_UP, 0x80, ""); err != nil {
			return err
		}
		if err := remote.SetKey(remotes.KEYCODE_VOLUME_DOWN, 0x00, ""); err != nil {
			return err
		}
		if err := remote.SetKey(remotes.KEYCODE_PLAY, 0x10, ""); err != nil {
			return err
		}

		// Save file
		if err := remote.SaveToFile(filename); err != nil {
			return err
		}
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
