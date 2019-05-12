/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2019
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/
package main

import (
	"os"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-hw/sys/lirc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Codecs
	_ "github.com/djthorpe/remotes/sys/sony"
)

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
	codecs := append(codecs())
	config := gopi.NewAppConfig(codecs...)

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
