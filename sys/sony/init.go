/*
	Go Language Raspberry Pi Interface
    (c) Copyright David Thorpe 2019
    All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package sony

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register remotes/sony12
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/sony12",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
				Type: remotes.CODEC_SONY12,
			}, app.Logger)
		},
	})

	// Register remotes/sony15
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/sony15",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
				Type: remotes.CODEC_SONY15,
			}, app.Logger)
		},
	})

	// Register remotes/sony20
	gopi.RegisterModule(gopi.Module{
		Name:     "remotes/sony20",
		Requires: []string{"lirc"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Codec{
				LIRC: app.ModuleInstance("lirc").(gopi.LIRC),
				Type: remotes.CODEC_SONY20,
			}, app.Logger)
		},
	})

}
