/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register remotes/keymap
	gopi.RegisterModule(gopi.Module{
		Name: "remotes/keymap",
		Type: gopi.MODULE_TYPE_KEYMAP,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("keymap.path", "", "Keymap File")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			path, _ := app.AppFlags.GetString("keymap.path")
			return gopi.Open(Keymap{
				Path: path,
			}, app.Logger)
		},
		Run: func(app *gopi.AppInstance, driver gopi.Driver) error {
			driver_ := driver.(*keymapper)
			for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_OTHER) {
				if codec, ok := app.ModuleInstance(module.Name).(remotes.Codec); ok {
					driver_.RegisterCodec(codec)
				}
			}
			return nil
		},
	})
}
