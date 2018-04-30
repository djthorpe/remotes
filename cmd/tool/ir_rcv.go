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
	"strings"
	"sync"
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
	_ "github.com/djthorpe/remotes/codec/nec"
	_ "github.com/djthorpe/remotes/codec/panasonic"
	_ "github.com/djthorpe/remotes/codec/rc5"
	_ "github.com/djthorpe/remotes/codec/sony"
)

var (
	start chan struct{}
	once  sync.Once
)

////////////////////////////////////////////////////////////////////////////////

func PrintHeader() {
	fmt.Printf("%-20s %-25v %-10s %-10s %-15s %-22s %s\n", "Name", "Key", "Scancode", "Device", "Codec", "Event", "Timestamp")
	fmt.Printf("%-20s %-25v %-10s %-10s %-15s %-22s %s\n",
		strings.Repeat("-", 20), strings.Repeat("-", 25), strings.Repeat("-", 10), strings.Repeat("-", 10),
		strings.Repeat("-", 15), strings.Repeat("-", 22), strings.Repeat("-", 10))
}

func PrintEntry(entry *remotes.KeyMapEntry, evt_type gopi.InputEventType, ts time.Duration) {
	once.Do(PrintHeader)
	ts = ts.Truncate(time.Millisecond)
	fmt.Printf("%-20s %-25v 0x%08X 0x%08X %-15s %-22s %v\n", entry.Name, entry.Keycode, entry.Scancode, entry.Device, entry.Type, evt_type, ts)
}

func PrintEvent(evt *remotes.RemoteEvent) {
	once.Do(PrintHeader)
	ts := evt.Timestamp().Truncate(time.Millisecond)
	fmt.Printf("%-20s %-25v 0x%08X 0x%08X %-15s %-22s %v\n", "<unmapped>", "<unmapped>", evt.Scancode(), evt.Device(), evt.Codec(), evt.EventType(), ts)
}

////////////////////////////////////////////////////////////////////////////////

func HandleEvent(keymaps remotes.KeyMaps, evt *remotes.RemoteEvent) error {
	// Lookup entry
	if entries := keymaps.LookupKeyMapEntry(evt.Codec(), evt.Device(), evt.Scancode()); entries != nil {
		for _, entry := range entries {
			PrintEntry(entry, evt.EventType(), evt.Timestamp())
		}
	} else {
		PrintEvent(evt)
	}
	return nil
}

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {

	app.Logger.Debug("Waiting for start signal")
	<-start
	app.Logger.Debug("Got start signal")

	// Output header
	once.Do(PrintHeader)

	// Create a merged event channel
	events := event.NewEventMerger()
	remote_events := events.Subscribe()

	// Subscribe to codecs
	for _, name := range codecs() {
		if instance, ok := app.ModuleInstance(name).(remotes.Codec); ok && instance != nil {
			events.Add(instance.Subscribe())
		}
	}

	// Obtain keymaps
	keymaps := app.ModuleInstance("keymap").(remotes.KeyMaps)

	// Wait for either terminate signal or incoming remote event
FOR_LOOP:
	for {
		select {
		case <-done:
			break FOR_LOOP
		case remote_event := <-remote_events:
			if err := HandleEvent(keymaps, remote_event.(*remotes.RemoteEvent)); err != nil {
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

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Load keymaps
	if keymaps := app.ModuleInstance("keymap").(remotes.KeyMaps); keymaps == nil {
		start <- gopi.DONE
		done <- gopi.DONE
		return errors.New("Missing keymaps")
	} else if err := keymaps.LoadKeyMaps(func(filename string, keymap *remotes.KeyMap) {
		app.Logger.Info("Loading: %v (%v)", filename, keymap.Name)
	}); err != nil {
		start <- gopi.DONE
		done <- gopi.DONE
		return err
	}

	// Start signal
	start <- gopi.DONE

	// Wait for interrupt
	app.Logger.Info("Waiting for CTRL+C or SIGTERM to end")
	app.WaitForSignal()

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
	codecs := append(codecs(), "remotes/keymap")
	config := gopi.NewAppConfig(codecs...)

	// start signal
	start = make(chan struct{})

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
