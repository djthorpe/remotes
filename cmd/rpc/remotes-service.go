/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
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
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"
	_ "github.com/djthorpe/gopi/sys/rpc/mdns"
	_ "github.com/djthorpe/remotes/keymap"

	// RPC Services
	_ "github.com/djthorpe/remotes/cmd/rpc/service"

	// Remote Codecs
	_ "github.com/djthorpe/remotes/codec/nec"
	_ "github.com/djthorpe/remotes/codec/panasonic"
	_ "github.com/djthorpe/remotes/codec/sony"
)

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
	// Create the configuration
	modules := append(codecs(), "service/remotes:grpc", "keymap")
	config := gopi.NewAppConfig(modules...)

	// Set the RPCServiceRecord for server discovery
	config.Service = "remotes"

	// Run the server and register all the services
	// Note the CommandLoop needs to go last as it blocks on Receive() until
	// Cancel is called from the CommandCancel task
	os.Exit(gopi.RPCServerTool(config))
}
