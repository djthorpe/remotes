/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

// ir_send is used to send learnt IR codes via LIRC
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/keymap"

	// Remotes
	_ "github.com/djthorpe/remotes/codec/appletv"
	_ "github.com/djthorpe/remotes/codec/nec"
	_ "github.com/djthorpe/remotes/codec/panasonic"
	_ "github.com/djthorpe/remotes/codec/sony"
)

////////////////////////////////////////////////////////////////////////////////

func DisplayKeymapsHeader() {
	fmt.Printf("%-20s %-25s %-17s %-10s %-10s %-7s\n", "KEY", "CODE", "CODEC", "DEVICE", "SCANCODE", "REPEATS")
	fmt.Printf("%-20s %-25s %-17s %-10s %-10s %-7s\n", strings.Repeat("-", 20), strings.Repeat("-", 25), strings.Repeat("-", 17), strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 7))
}

func DisplayEntryHeader() {
	fmt.Printf("%-20s %-25s %-17s %-10s %-10s %-7s\n", "KEY", "CODE", "CODEC", "DEVICE", "SCANCODE", "REPEATS")
	fmt.Printf("%-20s %-25s %-17s %-10s %-10s %-7s\n", strings.Repeat("-", 20), strings.Repeat("-", 25), strings.Repeat("-", 17), strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 7))
}

func DisplayKeymaps(keymaps remotes.KeyMaps, app *gopi.AppInstance) error {
	var once sync.Once
	var err error

	// Display all keymaps files
	for _, keymap := range keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, "") {
		once.Do(DisplayKeymapsHeader)
		fmt.Printf("%-20s %-20s 0x%08X %7d %7d\n", keymap.Name, keymap.Type, keymap.Device, len(keymap.Map), keymap.Repeats)
	}

	// If no keymaps displayed, then assign error
	once.Do(func() {
		err = fmt.Errorf("No keymaps to display")
	})

	// Return
	return err
}

func DisplayEntries(device string, keymaps remotes.KeyMaps, app *gopi.AppInstance) error {
	var once sync.Once
	var err error

	// Output all entries from device
	for _, keymap := range keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, device) {
		for _, entry := range keymaps.LookupKeyMapEntry(keymap.Type, keymap.Device, remotes.SCANCODE_UNKNOWN) {
			once.Do(DisplayEntryHeader)
			fmt.Printf("%-20s %-25s %-17v 0x%08X 0x%08X %7d\n", entry.Name, entry.Keycode, entry.Type, entry.Device, entry.Scancode, entry.Repeats)
		}
	}

	// If no keymaps displayed, then assign error
	once.Do(func() {
		err = fmt.Errorf("Invalid -device flag")
	})

	// Return
	return err
}

func Send(device string, keymaps remotes.KeyMaps, args []string) error {
	// Get keymap for device
	allkeymaps := keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, device)
	if len(allkeymaps) != 1 {
		return fmt.Errorf("Invalid -device flag")
	}

	// For each argument, return the set of keys and then match those keys
	// to a single entry or else the argument is ambiguous
	for _, arg := range args {
		entries := make([]*remotes.KeyMapEntry, 0, 1)
		for _, key := range keymaps.LookupKeyCode(arg) {
			entries = append(entries, keymaps.GetKeyMapEntry(allkeymaps[0], remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, key.Keycode, remotes.SCANCODE_UNKNOWN)...)
		}
		if len(entries) == 0 {
			return fmt.Errorf("Unknown key: %v", arg)
		} else if len(entries) > 1 {
			ambigious := ""
			for _, entry := range entries {
				ambigious += fmt.Sprint("'" + entry.Name + "',")
			}
			return fmt.Errorf("Ambiguous key: %v (It could mean one of %v)", arg, strings.TrimSuffix(ambigious, ","))
		} else {
			fmt.Printf("arg=%v entry=%v\n", arg, entries[0])
		}
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Load keymaps
	keymaps := app.ModuleInstance("keymap").(remotes.KeyMaps)
	if err := keymaps.LoadKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		app.Logger.Info("Loaded '%v' from file %v", keymap.Name, filename)
	}); err != nil {
		done <- gopi.DONE
		return err
	}

	if device, exists := app.AppFlags.GetString("device"); exists == false {
		if err := DisplayKeymaps(keymaps, app); err != nil {
			done <- gopi.DONE
			return err
		}
	} else if args := app.AppFlags.Args(); len(args) > 0 {
		// Send keycodes
		if err := Send(device, keymaps, args); err != nil {
			done <- gopi.DONE
			return err
		}
	} else if err := DisplayEntries(device, keymaps, app); err != nil {
		done <- gopi.DONE
		return err
	}

	// Wait for interrupt
	//app.Logger.Info("Waiting for CTRL+C or SIGTERM to end")
	//app.WaitForSignal()

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func codecs() []string {
	codecs := make([]string, 0)
	// Obtain all the codecs
	for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_OTHER) {
		if strings.HasPrefix(module.Name, "remotes/") {
			codecs = append(codecs, module.Name)
		}
	}
	return codecs
}

func main() {
	// Configuration
	codecs := append(codecs(), "keymap")
	config := gopi.NewAppConfig(codecs...)

	config.AppFlags.FlagString("device", "", "Name of device to send codes to")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
