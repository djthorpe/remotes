/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package appletv

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register remotes/appletv
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/appletv",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
			}, app.Logger)
		},
	})
}
