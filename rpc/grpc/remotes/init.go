/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	remotes "github.com/djthorpe/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register rpc/service/remotes:grpc
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/service/remotes:grpc",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "keymap"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Service{
				Server:  app.ModuleInstance("rpc/server").(gopi.RPCServer),
				KeyMaps: app.ModuleInstance("keymap").(remotes.KeyMaps),
			}, app.Logger)
		},
		Run: func(app *gopi.AppInstance, driver gopi.Driver) error {
			// Register codecs with driver. Codecs have OTHER as module type
			// and name starting with "remotes/"
			for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_OTHER) {
				if strings.HasPrefix(module.Name, "remotes/") {
					if codec, ok := app.ModuleInstance(module.Name).(remotes.Codec); ok && codec != nil {
						driver.(*service).registerCodec(codec)
					}
				}
			}
			// Load KeyMaps
			if err := driver.(*service).loadKeyMaps(); err != nil {
				return err
			}
			// Success
			return nil
		},
	})

	// Register rpc/client/remotes:grpc
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/client/remotes:grpc",
		Type:     gopi.MODULE_TYPE_CLIENT,
		Requires: []string{"rpc/clientpool"},
		Run: func(app *gopi.AppInstance, _ gopi.Driver) error {
			clientpool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
			if clientpool == nil {
				return gopi.ErrAppError
			} else {
				clientpool.RegisterClient("remotes.Remotes", NewClient)
				return nil
			}
		},
	})
}
