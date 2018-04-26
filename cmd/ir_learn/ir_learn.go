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
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/util/event"
	"github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/codec/appletv"
	_ "github.com/djthorpe/remotes/codec/nec"
	_ "github.com/djthorpe/remotes/codec/panasonic"
	_ "github.com/djthorpe/remotes/codec/sony"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type App struct {
	app      *gopi.AppInstance
	remote   *remotes.Remote
	key      *remotes.KeyMap // Currently learned key
	filename string
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS & GLOBALS

const (
	DEFAULT_EXT = ".keymap"
)

var (
	theApp      *App
	startSignal chan struct{}
)

////////////////////////////////////////////////////////////////////////////////
// New Application

func NewApp(app *gopi.AppInstance) *App {
	this := new(App)
	this.app = app
	this.remote = nil
	this.key = nil
	return this
}

////////////////////////////////////////////////////////////////////////////////
// App methods

func (this *App) Load(name string) error {

	// Set the name to include extension
	if filepath.Ext(name) != DEFAULT_EXT {
		name = name + DEFAULT_EXT
	}

	// Create or load from file
	if stat, err := os.Stat(name); os.IsNotExist(err) {
		device_name := filepath.Base(name)

		// Strip name of extension
		if strings.HasSuffix(device_name, DEFAULT_EXT) {
			device_name = strings.TrimSuffix(device_name, DEFAULT_EXT)
		}

		// Create a new remote
		if remote := remotes.NewRemote(remotes.CODEC_NONE, 0); remote == nil {
			return gopi.ErrBadParameter
		} else {
			remote.SetName(device_name)
			this.remote = remote
			this.filename = name
		}
	} else if stat.IsDir() {
		return gopi.ErrBadParameter
	} else if remote, err := remotes.RemoteFromFile(name); err != nil {
		return err
	} else {
		this.remote = remote
		this.filename = name
	}

	// Success
	return nil
}

func (this *App) SaveToFile() error {
	if this.remote == nil {
		return gopi.ErrOutOfOrder
	} else {
		return this.remote.SaveToFile(this.filename)
	}
}

func (this *App) SetKey(key *remotes.KeyMap) {
	this.key = key
}

func (this *App) HandleEvent(evt *remotes.RemoteEvent) error {
	if this.key != nil && evt != nil {
		// Set the codec if CODEC_NONE
		if this.remote.Type == remotes.CODEC_NONE {
			this.remote.Type = evt.Codec()
		} else if this.remote.Type != evt.Codec() {
			fmt.Printf("\n  Ignoring key, different codec (%v) than expected (%v)\n", evt.Codec(), this.remote.Type)
			return nil
		}

		// Check the device
		if this.remote.Device == 0 {
			this.remote.Device = evt.Device()
		} else if this.remote.Device != evt.Device() {
			fmt.Printf("\n  Ignoring key, different device (0x%08X) than expected (0x%08X)\n", evt.Device(), this.remote.Device)
			return nil
		}

		// Set the key
		if err := this.remote.SetKey(this.key.Keycode, evt.Scancode(), ""); err != nil {
			return err
		} else {
			fmt.Printf("\n  Learned device=%08X scancode=%08X => Key %v\n", evt.Device(), evt.Scancode(), this.key.Keycode)
		}
	}

	// Success
	return nil
}

// Keycodes returns the set of keys
func (this *App) Keycodes() ([]*remotes.KeyMap, error) {
	if keys, exists := this.app.AppFlags.GetString("key"); exists {
		keymaps := make([]*remotes.KeyMap, 0)
		for _, key := range strings.Split(keys, ",") {
			if k := remotes.KeycodesForString(key); k == nil {
				return nil, fmt.Errorf("Key(s) not found: %v", key)
			} else {
				keymaps = append(keymaps, k...)
			}
		}
		// TODO: Check for empty case
		// TODO: Check for duplicate keys
		return keymaps, nil
	} else {
		// Return all keycodes
		return remotes.Keycodes(), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// Main

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Set the application and signal start
	theApp = NewApp(app)
	startSignal <- gopi.DONE

	// If device is empty, then return error
	if device_name, exists := app.AppFlags.GetString("device"); exists == false {
		done <- gopi.DONE
		return fmt.Errorf("Missing -device flag")
	} else if err := theApp.Load(device_name); err != nil {
		done <- gopi.DONE
		return err
	} else {
		// Iterate through Keycodes
		if keycodes, err := theApp.Keycodes(); err != nil {
			done <- gopi.DONE
			return err
		} else {
			for i, key := range keycodes {
				fmt.Printf("(%v/%v) Press the \"%v\" key (%v) or wait for the next key...", i+1, len(keycodes), key.Name, key.Keycode)

				// Set the key we're currently learning
				theApp.SetKey(key)

				if app.WaitForSignalOrTimeout(5 * time.Second) {
					// Finish gracefully if signal caught
					done <- gopi.DONE
					fmt.Printf("\nTerminating without saving...\n")
					return nil
				}

				// Reset the key we're currently learning
				theApp.SetKey(nil)
				fmt.Println("")
			}
		}

		// Save file
		if err := theApp.SaveToFile(); err != nil {
			done <- gopi.DONE
			return err
		}
	}

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	// Wait for start signal
	<-startSignal

	// Create a merged event channel
	events := event.NewEventMerger()
	remote_events := events.Subscribe()

	// Subscribe to codecs
	for _, name := range codecs() {
		if instance, ok := app.ModuleInstance(name).(remotes.Codec); ok && instance != nil {
			events.Add(instance.Subscribe())
		}
	}

	// Wait for either terminate signal or incoming remote event
FOR_LOOP:
	for {
		select {
		case <-done:
			break FOR_LOOP
		case remote_event := <-remote_events:
			if err := theApp.HandleEvent(remote_event.(*remotes.RemoteEvent)); err != nil {
				app.Logger.Warn("EventLoop: %v", err)
			}
		}
	}

	// Close merged events
	events.Unsubscribe(remote_events)
	events.Close()
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
	// Obtain all codecs
	codecs := codecs()
	if len(codecs) == 0 {
		fmt.Fprintln(os.Stderr, "Missing codecs")
		os.Exit(-1)
	}

	// Configuration
	config := gopi.NewAppConfig(codecs...)
	config.AppFlags.FlagString("device", "", "Name of device to learn")
	config.AppFlags.FlagString("key", "", "Comma-separated list of keys to learn")

	// Set the start signal
	startSignal = make(chan struct{})

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
