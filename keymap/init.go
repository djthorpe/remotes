/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register remotes/appletv
	gopi.RegisterModule(gopi.Module{
		Name: "remotes/keymap",
		Type: gopi.MODULE_TYPE_KEYMAP,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("keymap.db", "/var/local/remotes", "Key mapping database path")
			config.AppFlags.FlagString("keymap.ext", "", "Key mapping file extension")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			root, _ := app.AppFlags.GetString("keymap.db")
			ext, _ := app.AppFlags.GetString("keymap.ext")
			return gopi.Open(Database{
				Root: root,
				Ext:  ext,
			}, app.Logger)
		},
	})
}
