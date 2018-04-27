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
	"strings"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/util/event"
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
// TYPES

type App struct {
	app    *gopi.AppInstance
	keymap *remotes.KeyMap      // Currently learned keymap
	key    *remotes.KeyMapEntry // Currently learned key
	db     remotes.KeyMaps
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	theApp      *App
	startSignal chan struct{}
)

////////////////////////////////////////////////////////////////////////////////
// New Application

func NewApp(app *gopi.AppInstance) *App {
	this := new(App)
	this.app = app
	this.key = nil
	this.keymap = nil
	this.db = app.ModuleInstance("keymap").(remotes.KeyMaps)

	// Load in the existing keymaps from root
	if err := this.db.LoadKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		app.Logger.Info("Loading: %v (%v)", filename, keymap.Name)
	}); err != nil {
		app.Logger.Error("Error: %v", err)
		return nil
	}

	// Return theApp
	return this
}

////////////////////////////////////////////////////////////////////////////////
// App methods

func (this *App) SetKey(keymap *remotes.KeyMap, key *remotes.KeyMapEntry) {
	this.keymap = keymap
	this.key = key
}

func (this *App) HandleEvent(evt *remotes.RemoteEvent) error {
	if this.keymap == nil && this.key == nil && evt == nil {
		return nil
	}

	if err := this.db.SetKeyMapEntry(this.keymap, evt.Codec(), evt.Device(), this.key.Keycode, evt.Scancode()); err != nil {
		fmt.Printf("\n  %v\n", err)
	} else {
		fmt.Printf("\n  Recorded key %v for device 0x%08X and scancode 0x%08X\n", this.key.Keycode, evt.Device(), evt.Scancode())
	}

	// Success
	return nil
}

// Return a keymap
func (this *App) KeyMapWithName(name string) *remotes.KeyMap {
	// Find keymaps with the name specified
	if keymaps := this.db.KeyMaps(remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN, name); len(keymaps) == 0 {
		this.app.Logger.Info("Creating a new keymap with name '%v'", name)
		return this.db.NewKeyMap(name)
	} else if len(keymaps) > 1 {
		this.app.Logger.Warn("Name '%v' matches more than one keymap, using the first one")
		return keymaps[0]
	} else {
		return keymaps[0]
	}
}

// Save modified keymaps
func (this *App) Save() error {
	return this.db.SaveModifiedKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		this.app.Logger.Info("Saving: %v (%v)", filename, keymap.Name)
	})
}

// Keycodes returns the set of keys
func (this *App) Keycodes() []*remotes.KeyMapEntry {
	if keys, exists := this.app.AppFlags.GetString("key"); exists {
		return this.db.LookupKeyCode(strings.Split(keys, ",")...)
	} else {
		return this.db.LookupKeyCode()
	}
}

////////////////////////////////////////////////////////////////////////////////
// Main

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Set the application and signal start
	theApp = NewApp(app)
	startSignal <- gopi.DONE
	if theApp == nil {
		done <- gopi.DONE
		return fmt.Errorf("Unable to create application")
	}

	// If device is empty, then return error
	if name, exists := app.AppFlags.GetString("device"); exists == false {
		done <- gopi.DONE
		return fmt.Errorf("Missing -device flag")
	} else if keymap := theApp.KeyMapWithName(name); keymap == nil {
		done <- gopi.DONE
		return gopi.ErrBadParameter
	} else {
		// Iterate through Keycodes
		keycodes := theApp.Keycodes()
		for i, key := range keycodes {
			fmt.Printf("(%v/%v) Press the \"%v\" key (%v) or wait for the next key...", i+1, len(keycodes), key.Name, key.Keycode)

			// Set the key we're currently learning
			theApp.SetKey(keymap, key)

			if app.WaitForSignalOrTimeout(5 * time.Second) {
				// Finish gracefully if signal caught
				done <- gopi.DONE
				fmt.Printf("\nTerminating without saving...\n")
				return nil
			}

			// Reset the key we're currently learning
			theApp.SetKey(keymap, nil)
			fmt.Println("")
		}
	}

	// Save any modified keymaps
	if err := theApp.Save(); err != nil {
		done <- gopi.DONE
		return err
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
	// Configuration
	codecs := append(codecs(), "keymap")
	config := gopi.NewAppConfig(codecs...)
	config.AppFlags.FlagString("device", "", "Name of device to learn")
	config.AppFlags.FlagString("key", "", "Comma-separated list of keys to learn")

	// Set the start signal
	startSignal = make(chan struct{})

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
