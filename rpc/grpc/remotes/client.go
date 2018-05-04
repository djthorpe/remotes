/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package remotes

import (
	"context"
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi/sys/rpc/grpc"
	remotes "github.com/djthorpe/remotes"

	// Protocol Buffer definitions
	pb "github.com/djthorpe/remotes/rpc/protobuf/remotes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.RemotesClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewRemotesClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

func (this *Client) NewContext() context.Context {
	if this.conn.Timeout() == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), this.conn.Timeout())
		return ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Client) Conn() gopi.RPCClientConn {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

// Return array of codecs supported
func (this *Client) Codecs() ([]remotes.CodecType, error) {
	ctx := this.NewContext()
	if reply, err := this.RemotesClient.Codecs(ctx, &pb.EmptyRequest{}); err != nil {
		return nil, err
	} else {
		codecs := make([]remotes.CodecType, len(reply.Codec))
		for i := range reply.Codec {
			codecs[i] = remotes.CodecType(reply.Codec[i])
		}
		return codecs, nil
	}
}

/*
// Return array of keymaps learnt
func (this *Client) KeyMaps() (*KeyMapsReply, error) {
	return nil, gopi.ErrNotImplemented
}

// Return keys learnt by keymap name
func (this *Client) Keys(in *KeysRequest) (*KeysReply, error) {
	return nil, gopi.ErrNotImplemented
}

// Return all possible keys with one or more search terms
func (this *Client) LookupKeys(in *LookupKeysRequest) (*KeysReply, error) {
	return nil, gopi.ErrNotImplemented
}

// Receive remote events
func (this *Client) Receive(in *EmptyRequest) (Remotes_ReceiveClient, error) {
	return nil, gopi.ErrNotImplemented
}

// Send a remote scancode
func (this *Client) SendScancode(in *SendScancodeRequest) error {
	return gopi.ErrNotImplemented
}

// Send a remote keycode
func (this *Client) SendKeycode(in *SendKeycodeRequest) error {
	return gopi.ErrNotImplemented
}
*/

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<grpc.client.remotes>{ conn=%v }", this.conn)
}
