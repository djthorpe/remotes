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
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/keymap"

	// Remotes
	_ "github.com/djthorpe/remotes/codec/nec"
	_ "github.com/djthorpe/remotes/codec/panasonic"
	_ "github.com/djthorpe/remotes/codec/sony"
)

////////////////////////////////////////////////////////////////////////////////

var (
	send chan *remotes.KeyMapEntry
)

////////////////////////////////////////////////////////////////////////////////

func DisplayKeymapsHeader() {
	fmt.Printf("%-20s %-20s %-10s %-7s %-7s\n", "DEVICE", "CODEC", "ID", "KEYS", "REPEATS")
	fmt.Printf("%-20s %-20s %-10s %-7s %-7s\n", strings.Repeat("-", 20), strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 7), strings.Repeat("-", 7))
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

func DisplayEntry(entry *remotes.KeyMapEntry) {
	fmt.Printf("%-20s %-25s %-17v 0x%08X 0x%08X %7d\n", entry.Name, entry.Keycode, entry.Type, entry.Device, entry.Scancode, entry.Repeats)
}

func DisplayDeviceEntries(device string, keymaps remotes.KeyMaps, app *gopi.AppInstance) error {
	var once sync.Once
	var err error

	// Output all entries from device
	for _, keymap := range keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, device) {
		for _, entry := range keymaps.LookupKeyMapEntry(keymap.Type, keymap.Device, remotes.SCANCODE_UNKNOWN) {
			once.Do(DisplayEntryHeader)
			DisplayEntry(entry)
		}
	}

	// If no keymaps displayed, then assign error
	once.Do(func() {
		err = fmt.Errorf("Invalid -device flag")
	})

	// Return
	return err
}

func Send(device string, keymaps remotes.KeyMaps, args []string, repeats uint, repeats_override bool) error {
	var once sync.Once

	// Get keymap for device
	allkeymaps := keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, device)
	if len(allkeymaps) != 1 {
		return fmt.Errorf("Invalid -device flag")
	}

	// For each argument, return the set of keys and then match those keys
	// to a single entry or else the argument is ambiguous
	allkeys := strings.Split(strings.Join(args, ","), ",")
	fmt.Println(allkeys)
	for _, arg := range allkeys {
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
			// Override the repeats value
			if repeats_override {
				entries[0].Repeats = repeats
			}
			// Perform the send
			once.Do(DisplayEntryHeader)
			DisplayEntry(entries[0])
			send <- entries[0]
		}
	}

	// Return success
	return nil
}

func SetRepeats(device string, keymaps remotes.KeyMaps, repeats uint) error {
	// Get keymap for device
	allkeymaps := keymaps.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, device)
	if len(allkeymaps) != 1 {
		return fmt.Errorf("Invalid -device flag")
	}
	// Set the parameter
	return keymaps.SetRepeats(allkeymaps[0], repeats)
}

////////////////////////////////////////////////////////////////////////////////

func SendLoop(app *gopi.AppInstance, done <-chan struct{}) error {

	// Make a map of the codecs registered
	codec_map := make(map[remotes.CodecType]remotes.Codec, 10)
	for _, name := range codecs() {
		if instance, ok := app.ModuleInstance(name).(remotes.Codec); ok && instance != nil {
			codec_map[instance.Type()] = instance
		}
	}

FOR_LOOP:
	for {
		select {
		case <-done:
			break FOR_LOOP
		case entry := <-send:
			if entry != nil {
				if codec, exists := codec_map[entry.Type]; exists == false || codec == nil {
					return fmt.Errorf("Codec not registered: %v", entry.Type)
				} else if err := codec.Send(entry.Device, entry.Scancode, entry.Repeats); err != nil {
					return err
				}
			}
		}
	}

	// Success
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Load keymaps
	keymaps := app.ModuleInstance("keymap").(remotes.KeyMaps)
	if err := keymaps.LoadKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		app.Logger.Info("Loaded '%v' from file %v", keymap.Name, filename)
	}); err != nil {
		done <- gopi.DONE
		return err
	}

	// Repeats override
	repeats, repeats_override := app.AppFlags.GetUint("repeats")

	if device, exists := app.AppFlags.GetString("device"); exists == false {
		// No -device flag so display devices
		if err := DisplayKeymaps(keymaps, app); err != nil {
			done <- gopi.DONE
			return err
		}
	} else if args := app.AppFlags.Args(); len(args) > 0 {
		// Send keycodes
		if err := Send(device, keymaps, args, repeats, repeats_override); err != nil {
			done <- gopi.DONE
			return err
		}
	} else {
		if repeats_override {
			// Set repeats value
			if err := SetRepeats(device, keymaps, repeats); err != nil {
				done <- gopi.DONE
				return err
			}
		}
		if err := DisplayDeviceEntries(device, keymaps, app); err != nil {
			done <- gopi.DONE
			return err
		}
	}

	// Wait for interrupt
	app.Logger.Info("Waiting for CTRL+C or SIGTERM to end")
	app.WaitForSignalOrTimeout(500 * time.Millisecond)

	// Save
	if err := keymaps.SaveModifiedKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		app.Logger.Info("Saving '%v' to file %v", keymap.Name, filename)
	}); err != nil {
		done <- gopi.DONE
		return err
	}

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
	config.AppFlags.FlagUint("repeats", 0, "Number of code repeats (overrides default)")

	// Make the send channel
	send = make(chan *remotes.KeyMapEntry)

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, SendLoop))
}
