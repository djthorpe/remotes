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
	"strings"
	"sync"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/remotes"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"
	_ "github.com/djthorpe/gopi/sys/rpc/mdns"

	// RPC Client
	client "github.com/djthorpe/remotes/rpc/grpc/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	EventChannel    = make(chan *client.Event)
	PrintHeaderOnce sync.Once
)

////////////////////////////////////////////////////////////////////////////////
// STRING FORMATTING

func fmtRepeats(repeats uint) string {
	if repeats > 0 {
		return fmt.Sprint(repeats)
	} else {
		return ""
	}
}

func fmtDevice(device uint32) string {
	if device == remotes.DEVICE_UNKNOWN || device == 0 {
		return ""
	} else {
		return fmt.Sprintf("0x%08X", device)
	}
}

func fmtScancode(scancode uint32) string {
	if scancode == remotes.SCANCODE_UNKNOWN {
		return ""
	} else {
		return fmt.Sprintf("0x%X", scancode)
	}
}

func fmtCodec(codec remotes.CodecType) string {
	if codec == remotes.CODEC_NONE {
		return ""
	} else {
		return fmt.Sprint(codec)
	}
}

func fmtKey(key client.Key) string {
	return fmt.Sprintf("%v [%v]", key.Name, key.Keycode)
}

func fmtKeycode(keycode remotes.RemoteCode) string {
	return fmt.Sprintf("%d", keycode)
}

func fmtTimestamp(ts time.Duration) string {
	ts = ts.Truncate(time.Millisecond)
	return fmt.Sprint(ts)
}

func receivePrintHeader() {
	fmt.Printf("%-30s %-25s %-10s %-10s %-10s %-10s\n", "Key", "Event", "Keymap", "Device", "Scancode", "Timestamp")
	fmt.Printf("%-30s %-25s %-10s %-10s %-10s %-10s\n", "-------------------", "-------------------------", "----------", "----------", "----------", "----------")
}

func receivePrintEvent(event *client.Event) {
	PrintHeaderOnce.Do(receivePrintHeader)
	fmt.Printf("%-30s %-25s %-10s %-10s %-10s %-10s\n", fmtKey(event.Key), event.InputEvent.EventType, event.KeyMapInfo.Name, fmtDevice(event.Key.Device), fmtScancode(event.Key.Scancode), fmtTimestamp(event.InputEvent.Timestamp))
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT OPERATIONS

func Codecs(app *gopi.AppInstance, client *client.Client) error {
	if codecs, err := client.Codecs(); err != nil {
		return err
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Codec"})
		for _, codec := range codecs {
			table.Append([]string{
				fmtCodec(codec),
			})
		}
		table.Render()
	}
	return nil
}

func KeyMaps(app *gopi.AppInstance, client *client.Client) error {
	if keymaps, err := client.KeyMaps(); err != nil {
		return err
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Keymap", "Keys", "Codec", "Device", "Repeats"})
		for _, keymap := range keymaps {
			table.Append([]string{
				keymap.Name,
				fmt.Sprint(keymap.Keys),
				fmtCodec(keymap.Type),
				fmtDevice(keymap.Device),
				fmtRepeats(keymap.Repeats),
			})
		}
		table.Render()
	}
	return nil
}

func Keys(app *gopi.AppInstance, client *client.Client) error {
	keymap, _ := app.AppFlags.GetString("keymap")
	if keys, err := client.Keys(keymap); err != nil {
		return err
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Key", "Keycode", "Device", "Scancode", "Codec", "Repeats"})
		for _, key := range keys {
			table.Append([]string{
				key.Name,
				fmt.Sprint(key.Keycode),
				fmtDevice(key.Device),
				fmtScancode(key.Scancode),
				fmtCodec(key.Type),
				fmtRepeats(key.Repeats),
			})
		}
		table.Render()
	}
	return nil
}

func Receive(app *gopi.AppInstance, client *client.Client) error {

	// Make a channel to receive error on and the context
	errchan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	// Receive in background until cancel
	go func() {
		errchan <- client.Receive(ctx, EventChannel)
	}()

	fmt.Println("Press CTRL+C to stop receiving events")
	app.WaitForSignal()

	// Cancel, retrieve error and return
	cancel()
	err := <-errchan
	return err
}

func LookupKeys(app *gopi.AppInstance, client *client.Client) error {
	keymap, _ := app.AppFlags.GetString("keymap")
	terms := strings.Split(strings.Join(app.AppFlags.Args(), ","), ",")
	send, _ := app.AppFlags.GetBool("send")
	repeats, _ := app.AppFlags.GetUint("repeats")

	if keys, err := client.LookupKeys(keymap, terms); err != nil {
		return err
	} else if send && len(keys) > 1 {
		// Ambigous key
		return remotes.ErrAmbiguous
	} else if send && len(keys) == 1 {
		// Send key
		if err := client.SendKeycode(keymap, keys[0].Keycode, repeats); err != nil {
			return err
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Key", "Keycode", "Device", "Scancode", "Codec", "Repeats"})
		for _, key := range keys {
			table.Append([]string{
				"Sent: " + key.Name,
				fmt.Sprint(key.Keycode),
				fmtDevice(key.Device),
				fmtScancode(key.Scancode),
				fmtCodec(key.Type),
				fmtRepeats(key.Repeats),
			})
		}
		table.Render()
	} else if len(keys) == 0 {
		return gopi.ErrNotFound
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Key", "Keycode", "Device", "Scancode", "Codec", "Repeats"})
		for _, key := range keys {
			table.Append([]string{
				key.Name,
				fmt.Sprint(key.Keycode),
				fmtDevice(key.Device),
				fmtScancode(key.Scancode),
				fmtCodec(key.Type),
				fmtRepeats(key.Repeats),
			})
		}
		table.Render()
	}
	return nil
}

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

func GetClient(app *gopi.AppInstance) (*client.Client, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := GetConnection(app); err != nil {
		return nil, err
	} else if client_ := pool.NewClient("remotes.Remotes", conn); client_ == nil {
		return nil, gopi.ErrAppError
	} else if client, ok := client_.(*client.Client); ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client, nil
	}
}

func ReceiveLoop(app *gopi.AppInstance, done <-chan struct{}) error {
FOR_LOOP:
	for {
		select {
		case <-done:
			break FOR_LOOP
		case input := <-EventChannel:
			receivePrintEvent(input)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if client, err := GetClient(app); err != nil {
		done <- gopi.DONE
		return err
	} else {
		if len(app.AppFlags.Args()) == 0 {
			if _, exists := app.AppFlags.GetString("keymap"); exists {
				if err := Keys(app, client); err != nil {
					done <- gopi.DONE
					return err
				}
			} else {
				if err := Codecs(app, client); err != nil {
					done <- gopi.DONE
					return err
				}
				if err := KeyMaps(app, client); err != nil {
					done <- gopi.DONE
					return err
				}
				if err := Keys(app, client); err != nil {
					done <- gopi.DONE
					return err
				}
				if err := Receive(app, client); err != nil {
					done <- gopi.DONE
					return err
				}
			}
		} else {
			if err := LookupKeys(app, client); err != nil {
				done <- gopi.DONE
				return err
			}
		}
	}

	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/client/remotes:grpc")
	config.AppFlags.FlagString("addr", "", "Gateway address")
	config.AppFlags.FlagString("keymap", "", "Keymap")
	config.AppFlags.FlagBool("send", false, "Send keycode")
	config.AppFlags.FlagUint("repeats", 0, "Override repeats value")

	// Set the RPCServiceRecord for server discovery
	config.Service = "remotes"

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, ReceiveLoop))
}
