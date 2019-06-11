/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2019
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi-hw/sys/lirc"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/remotes/sys/keymap"

	// Codecs
	//_ "github.com/djthorpe/remotes/sys/nec"
	_ "github.com/djthorpe/remotes/sys/panasonic"
	//_ "github.com/djthorpe/remotes/sys/rc5"
	//_ "github.com/djthorpe/remotes/sys/sony"
)

var (
	once sync.Once
)

func PrintEvent(evt remotes.RemoteEvent) {
	once.Do(func() {
		fmt.Printf("SCAN DEVICE   CODEC      REPEAT\n")
	})
	codec := strings.TrimLeft(fmt.Sprint(evt.Codec()), "CODEC_")
	repeat := strings.TrimLeft(fmt.Sprint(evt.EventType()), "INPUT_EVENT_")
	fmt.Printf("0x%02X 0x%06X %-10s %-10s\n", evt.ScanCode(), evt.Device(), codec, repeat)

}

func Receive(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {

	// Set a longish LIRC timeout value
	if err := app.LIRC.SetRcvTimeout(1000 * 200); err != nil {
		return err
	}

	// Create generic keymap database
	db := app.ModuleInstance("keymap").(remotes.KeymapDatabase)
	keymap, err := db.New("test", remotes.CODEC_NONE, remotes.DEVICE_UNKNOWN)
	if err != nil && err != remotes.ErrDuplicateKeyMap {
		return err
	}

	// Subsribe to keymapper
	evt := db.Subscribe()

	// Start event loop
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case <-stop:
			break FOR_LOOP
		case event := <-evt:
			if event != nil {
				PrintEvent(event.(remotes.RemoteEvent))

				// Map key
				if key, err := db.Map(keymap, event.(remotes.RemoteEvent), gopi.KEYCODE_0, gopi.KEYSTATE_NONE); err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Mapped:", key)
				}
			}
		}
	}

	// Stop background task
	db.Unsubscribe(evt)

	// Return success
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	app.Logger.Info("Press CTRL+C to end")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func modules() []string {
	codecs := make([]string, 0)
	// Obtain the codecs
	for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_OTHER) {
		if strings.HasPrefix(module.Name, "remotes/") {
			codecs = append(codecs, module.Name)
		}
	}
	// Include keymapper
	for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_KEYMAP) {
		if strings.HasPrefix(module.Name, "remotes/") {
			codecs = append(codecs, module.Name)
		}
	}
	return codecs
}

func main() {
	// Append the modules and the keymapper to the configuration
	config := gopi.NewAppConfig(modules()...)

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Receive))
}
