/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2016-2018
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package nec

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register remotes/nec32
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/nec32",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
				Type: remotes.CODEC_NEC32,
			}, app.Logger)
		},
	})

	// Register remotes/nec16
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/nec16",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
				Type: remotes.CODEC_NEC16,
			}, app.Logger)
		},
	})

	// Register remotes/appletv
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/appletv2",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
				Type: remotes.CODEC_APPLETV,
			}, app.Logger)
		},
	})

}
