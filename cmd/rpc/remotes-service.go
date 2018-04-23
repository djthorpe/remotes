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

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/linux"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"
	_ "github.com/djthorpe/gopi/sys/rpc/mdns"

	// RPC Services
	_ "github.com/djthorpe/remotes/cmd/rpc/service"

	// Remote Codecs
	_ "github.com/djthorpe/remotes/codec/appletv"
	_ "github.com/djthorpe/remotes/codec/nec"
	_ "github.com/djthorpe/remotes/codec/panasonic"
	_ "github.com/djthorpe/remotes/codec/sony"
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("service/remotes:grpc", "remotes/appletv", "remotes/panasonic", "remotes/nec16", "remotes/nec32", "remotes/sony12", "remotes/sony15")

	// Set the RPCServiceRecord for server discovery
	config.Service = "remotes"

	// Run the server and register all the services
	// Note the CommandLoop needs to go last as it blocks on Receive() until
	// Cancel is called from the CommandCancel task
	os.Exit(gopi.RPCServerTool(config))
}
