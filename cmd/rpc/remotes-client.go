/*
   Go Language Raspberry Pi Interface
   (c) Copyright David Thorpe 2016-2018
   All Rights Reserved
   Documentation http://djthorpe.github.io/gopi/
   For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"syscall"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/protobuf/remotes"
	ptype "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MethodFunc func(app *gopi.AppInstance, service pb.RemotesClient, done <-chan struct{}) error

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	start  chan pb.RemotesClient
	method MethodFunc
)

var (
	methods = map[string]MethodFunc{
		"receive": Receive,
		"codecs":  Codecs,
		"keymaps": Keymaps,
		"keys":    Keys,
		"send":    Send,
	}
)

////////////////////////////////////////////////////////////////////////////////
// MAIN

func HasService(services []string, service string) bool {
	if services == nil {
		return false
	}
	for _, v := range services {
		if v == service {
			return true
		}
	}
	return false
}

func RemotesService(app *gopi.AppInstance) (pb.RemotesClient, error) {
	if client := app.ModuleInstance("rpc/clientconn").(gopi.RPCClientConn); client == nil {
		return nil, fmt.Errorf("Missing rpc/clientconn module")
	} else if services, err := client.Connect(); err != nil {
		return nil, err
	} else if has_service := HasService(services, "mutablelogic.Remotes"); has_service == false {
		return nil, fmt.Errorf("Missing Remotes service, gateway has the following services: %v", strings.Join(services, ","))
	} else if obj, err := client.NewService(reflect.ValueOf(pb.NewRemotesClient)); err != nil {
		return nil, err
	} else if service, ok := obj.(pb.RemotesClient); service == nil || ok == false {
		return nil, errors.New("Invalid remotes service")
	} else {
		return service, nil
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// There should be zero or one argument
	if args := app.AppFlags.Args(); len(args) > 1 {
		done <- gopi.DONE
		return gopi.ErrHelp
	} else if len(args) == 0 {
		method = Receive
	} else if method_, exists := methods[args[0]]; exists == false {
		done <- gopi.DONE
		return gopi.ErrBadParameter
	} else {
		method = method_
	}

	if service, err := RemotesService(app); err != nil {
		done <- gopi.DONE
		return err
	} else {
		// Initiate the service
		start <- service
	}

	// Wait for signal
	app.Logger.Info("Waiting for CTRL+C")
	app.WaitForSignal()

	// Finish gracefully
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Run RPC methods

func PrintEvent(once *sync.Once, reply *pb.ReceiveReply) {
	once.Do(func() {
		fmt.Printf("%-20s %-25s %-10s %-10s %-10s\n", "Codec", "Event", "Device", "Scancode", "Timestamp")
		fmt.Printf("%-20s %-25s %-10s %-10s %-10s\n", "-------------------", "-------------------------", "----------", "----------", "----------")
	})
	duration, _ := ptype.Duration(reply.Event.Ts)
	fmt.Printf("%-20s %-25s 0x%08X 0x%08X %-10s\n", reply.Key, reply.Event.EventType, reply.Event.Device, reply.Event.Scancode, duration)
}

func PrintCodecs(reply *pb.CodecsReply) {
	fmt.Printf("%-20s\n", "Codec")
	fmt.Printf("%-20s\n", "-------------------")
	for _, v := range reply.Codec {
		fmt.Printf("%-20s\n", v)
	}
}

func PrintKeymaps(reply *pb.KeyMapsReply) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Keymap", "Codec", "Device", "Repeats", "Keys"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	for _, key := range reply.Keymap {
		table.Append([]string{
			key.Name,
			fmt.Sprintf("%s", key.Codec),
			fmt.Sprintf("%X", key.Device),
			fmt.Sprintf("%d", key.Repeats),
			fmt.Sprintf("%d", key.Keys),
		})
	}
	table.Render()
}

func PrintKeys(reply *pb.KeysReply) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Code", "Codec", "Device", "Scancode", "Repeats"})
	table.SetAutoMergeCells(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	sort.Sort(keysByName(reply.Key))
	for _, key := range reply.Key {
		table.Append([]string{
			key.Name,
			key.Keycode,
			fmt.Sprintf("%s", key.Codec),
			fmt.Sprintf("%X", key.Device),
			fmt.Sprintf("%X", key.Scancode),
			fmt.Sprintf("%d", key.Repeats),
		})
	}
	table.Render()
}

type keysByName []*pb.Key

func (k keysByName) Len() int           { return len(k) }
func (k keysByName) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }
func (k keysByName) Less(i, j int) bool { return strings.Compare(k[i].Name, k[j].Name) == -1 }

func Context(app *gopi.AppInstance) (context.Context, context.CancelFunc) {
	if timeout, _ := app.AppFlags.GetDuration("rpc.timeout"); timeout == 0 {
		return context.WithCancel(context.Background())
	} else {
		return context.WithTimeout(context.Background(), timeout)
	}
}

func Receive(app *gopi.AppInstance, service pb.RemotesClient, done <-chan struct{}) error {
	var once sync.Once

	// Create the context with cancel
	ctx, cancel := Context(app)

	// Call cancel in the background when done is received
	go func() {
		<-done
		cancel()
	}()

	// Receive a stream of input events from the server
	if stream, err := service.Receive(ctx, &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		for {
			if input_event, err := stream.Recv(); err == io.EOF {
				break
			} else if err != nil {
				return err
			} else {
				PrintEvent(&once, input_event)
			}
		}
	}

	// Success
	return nil
}

func Codecs(app *gopi.AppInstance, service pb.RemotesClient, done <-chan struct{}) error {
	// Create the context with cancel
	ctx, cancel := Context(app)

	// Call cancel in the background when done is received
	go func() {
		<-done
		cancel()
	}()

	if reply, err := service.Codecs(ctx, &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		PrintCodecs(reply)
		return nil
	}
}

func Keymaps(app *gopi.AppInstance, service pb.RemotesClient, done <-chan struct{}) error {
	// Create the context with cancel
	ctx, cancel := Context(app)

	// Call cancel in the background when done is received
	go func() {
		<-done
		cancel()
	}()

	if reply, err := service.KeyMaps(ctx, &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		PrintKeymaps(reply)
		return nil
	}
}

func Keys(app *gopi.AppInstance, service pb.RemotesClient, done <-chan struct{}) error {
	// Create the context with cancel
	ctx, cancel := Context(app)

	// Call cancel in the background when done is received
	go func() {
		<-done
		cancel()
	}()

	if keymap, exists := app.AppFlags.GetString("keymap"); exists == false || keymap == "" {
		if reply, err := service.LookupKeys(ctx, &pb.LookupKeysRequest{}); err != nil {
			return err
		} else {
			PrintKeys(reply)
			return nil
		}
	} else {
		if reply, err := service.Keys(ctx, &pb.KeysRequest{Keymap: keymap}); err != nil {
			return err
		} else {
			PrintKeys(reply)
			return nil
		}
	}
}

func Send(app *gopi.AppInstance, service pb.RemotesClient, done <-chan struct{}) error {
	// Create the context with cancel
	ctx, cancel := Context(app)

	// Call cancel in the background when done is received
	go func() {
		<-done
		cancel()
	}()

	// Send can be:
	//   -keymap "AppleTV" -repeats <n> send <key>
	// or
	//   -codec <codec> -repeats <n> send <device> <scancode>

	/* Get command line arguments */
	if codec, exists := app.AppFlags.GetString("codec"); exists == false {
		return fmt.Errorf("Missing -codec argument")
	} else {
		codec = "CODEC_" + strings.ToUpper(codec)
		device, _ := app.AppFlags.GetUint("device")
		scancode, _ := app.AppFlags.GetUint("code")
		repeats, _ := app.AppFlags.GetUint("repeats")
		if codec_value, exists := pb.CodecType_value[codec]; exists == false {
			return fmt.Errorf("Unknown codec: %v", codec)
		} else {
			app.Logger.Debug("Send codec=%v device=0x%08X scancode=0x%08X repeats=%v", pb.CodecType(codec_value), device, scancode, repeats)
			if _, err := service.SendScancode(ctx, &pb.SendScancodeRequest{
				Codec:    pb.CodecType(codec_value),
				Device:   uint32(device),
				Scancode: uint32(scancode),
				Repeats:  uint32(repeats),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Run method and cancel main thread when done

func Cancel() error {
	if process, err := os.FindProcess(os.Getpid()); err != nil {
		return err
	} else if err := process.Signal(syscall.SIGTERM); err != nil {
		return err
	} else {
		return nil
	}
}

func Run(app *gopi.AppInstance, done <-chan struct{}) error {

FOR_LOOP:
	for {
		select {
		case service := <-start:
			if err := method(app, service, done); err != nil {
				Cancel()
				return err
			} else {
				Cancel()
				break FOR_LOOP
			}
		case <-done:
			break FOR_LOOP
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Bootstrap

func MethodNames() []string {
	v := make([]string, 0, len(methods))
	for k := range methods {
		v = append(v, k)
	}
	return v
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/clientconn")

	// Set send flags
	config.AppFlags.FlagString("keymap", "", "Keymap name")
	config.AppFlags.FlagString("codec", "", "Send Codec")
	config.AppFlags.FlagUint("device", 0, "Send device")
	config.AppFlags.FlagUint("code", 0, "Send scancode")
	config.AppFlags.FlagUint("repeats", 0, "Number of send repetitions")

	// Set usage function
	config.AppFlags.SetUsageFunc(func(flags *gopi.Flags) {
		fmt.Fprintf(os.Stderr, "Syntax:\n")
		fmt.Fprintf(os.Stderr, "  %v <flags> (%v)\n", flags.Name(), strings.Join(MethodNames(), "|"))
		fmt.Fprintf(os.Stderr, "\nWhere flags are:\n")
		flags.PrintDefaults()
	})

	// Create a start channel used to pass the service on indicating
	// the start of processing
	start = make(chan pb.RemotesClient)

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, Run))
}
