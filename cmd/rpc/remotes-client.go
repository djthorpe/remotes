/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

// RPC client for the the remotes-server
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes/rpc/grpc/remotes"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"
	_ "github.com/djthorpe/gopi/sys/rpc/mdns"
)

////////////////////////////////////////////////////////////////////////////////

func GetConnection(app *gopi.AppInstance) (gopi.RPCClientConn, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	addr, _ := app.AppFlags.GetString("addr")
	timeout, _ := app.AppFlags.GetDuration("rpc.timeout")
	// Have *some* time to lookup services. In fact, we should use
	// infinite timeout with 0 and have the lookup cancel when CTRL+C
	// it pressed - requires using a background task
	if timeout == 0 {
		timeout = 500 * time.Millisecond
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	if records, err := pool.Lookup(ctx, "", addr, 1); err != nil {
		return nil, err
	} else if len(records) == 0 {
		return nil, gopi.ErrDeadlineExceeded
	} else if conn, err := pool.Connect(records[0], 0); err != nil {
		return nil, err
	} else if services, err := conn.Services(); err != nil {
		return nil, err
	} else {
		app.Logger.Info("conn=%v services=%v", conn, services)
		return conn, nil
	}
}

func GetClient(app *gopi.AppInstance) (*remotes.Client, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := GetConnection(app); err != nil {
		return nil, err
	} else if client_ := pool.NewClient("remotes.Remotes", conn); client_ == nil {
		return nil, gopi.ErrAppError
	} else if client, ok := client_.(*remotes.Client); ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client, nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if client, err := GetClient(app); err != nil {
		done <- gopi.DONE
		return err
	} else if codecs, err := client.Codecs(); err != nil {
		done <- gopi.DONE
		return err
	} else {
		fmt.Printf("codecs=%v", codecs)
	}
	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/client/remotes:grpc")
	config.AppFlags.FlagString("addr", "", "Gateway address")

	// Set the RPCServiceRecord for server discovery
	config.Service = "remotes"

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
