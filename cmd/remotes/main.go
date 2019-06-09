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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	event "github.com/djthorpe/gopi/util/event"
	remotes "github.com/djthorpe/remotes"

	// Modules
	_ "github.com/djthorpe/gopi-hw/sys/lirc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Codecs
	_ "github.com/djthorpe/remotes/sys/nec"
	_ "github.com/djthorpe/remotes/sys/sony"
)

func Receive(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	// Create a merged event channel
	merger := event.Merger{}

	// Get codecs
	for _, codec := range codecs() {
		if instance, ok := app.ModuleInstance(codec).(remotes.Codec); ok && instance != nil {
			merger.Merge(instance)
		}
	}

	// Subsribe to merger
	evt := merger.Subscribe()

	// Start event loop
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case <-stop:
			break FOR_LOOP
		case event := <-evt:
			fmt.Println(event)
		}
	}

	// Stop background task
	merger.Unsubscribe(evt)
	merger.Close()

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
	// Append the codecs
	config := gopi.NewAppConfig(codecs()...)

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Receive))
}
