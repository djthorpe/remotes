// Code generated by protoc-gen-go. DO NOT EDIT.
// source: remotes/remotes.proto

/*
Package remotes is a generated protocol buffer package.

It is generated from these files:
	remotes/remotes.proto

It has these top-level messages:
	EmptyRequest
	EmptyReply
	Point
	InputEvent
	ReceiveReply
	SendScancodeRequest
	CodecsReply
*/
package remotes

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/duration"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type InputDeviceType int32

const (
	InputDeviceType_INPUT_TYPE_NONE        InputDeviceType = 0
	InputDeviceType_INPUT_TYPE_KEYBOARD    InputDeviceType = 1
	InputDeviceType_INPUT_TYPE_MOUSE       InputDeviceType = 2
	InputDeviceType_INPUT_TYPE_TOUCHSCREEN InputDeviceType = 4
	InputDeviceType_INPUT_TYPE_JOYSTICK    InputDeviceType = 8
	InputDeviceType_INPUT_TYPE_REMOTE      InputDeviceType = 16
)

var InputDeviceType_name = map[int32]string{
	0:  "INPUT_TYPE_NONE",
	1:  "INPUT_TYPE_KEYBOARD",
	2:  "INPUT_TYPE_MOUSE",
	4:  "INPUT_TYPE_TOUCHSCREEN",
	8:  "INPUT_TYPE_JOYSTICK",
	16: "INPUT_TYPE_REMOTE",
}
var InputDeviceType_value = map[string]int32{
	"INPUT_TYPE_NONE":        0,
	"INPUT_TYPE_KEYBOARD":    1,
	"INPUT_TYPE_MOUSE":       2,
	"INPUT_TYPE_TOUCHSCREEN": 4,
	"INPUT_TYPE_JOYSTICK":    8,
	"INPUT_TYPE_REMOTE":      16,
}

func (x InputDeviceType) String() string {
	return proto.EnumName(InputDeviceType_name, int32(x))
}
func (InputDeviceType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type InputEventType int32

const (
	InputEventType_INPUT_EVENT_NONE          InputEventType = 0
	InputEventType_INPUT_EVENT_KEYPRESS      InputEventType = 1
	InputEventType_INPUT_EVENT_KEYRELEASE    InputEventType = 2
	InputEventType_INPUT_EVENT_KEYREPEAT     InputEventType = 3
	InputEventType_INPUT_EVENT_ABSPOSITION   InputEventType = 4
	InputEventType_INPUT_EVENT_RELPOSITION   InputEventType = 5
	InputEventType_INPUT_EVENT_TOUCHPRESS    InputEventType = 6
	InputEventType_INPUT_EVENT_TOUCHRELEASE  InputEventType = 7
	InputEventType_INPUT_EVENT_TOUCHPOSITION InputEventType = 8
)

var InputEventType_name = map[int32]string{
	0: "INPUT_EVENT_NONE",
	1: "INPUT_EVENT_KEYPRESS",
	2: "INPUT_EVENT_KEYRELEASE",
	3: "INPUT_EVENT_KEYREPEAT",
	4: "INPUT_EVENT_ABSPOSITION",
	5: "INPUT_EVENT_RELPOSITION",
	6: "INPUT_EVENT_TOUCHPRESS",
	7: "INPUT_EVENT_TOUCHRELEASE",
	8: "INPUT_EVENT_TOUCHPOSITION",
}
var InputEventType_value = map[string]int32{
	"INPUT_EVENT_NONE":          0,
	"INPUT_EVENT_KEYPRESS":      1,
	"INPUT_EVENT_KEYRELEASE":    2,
	"INPUT_EVENT_KEYREPEAT":     3,
	"INPUT_EVENT_ABSPOSITION":   4,
	"INPUT_EVENT_RELPOSITION":   5,
	"INPUT_EVENT_TOUCHPRESS":    6,
	"INPUT_EVENT_TOUCHRELEASE":  7,
	"INPUT_EVENT_TOUCHPOSITION": 8,
}

func (x InputEventType) String() string {
	return proto.EnumName(InputEventType_name, int32(x))
}
func (InputEventType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type CodecType int32

const (
	CodecType_CODEC_RC5       CodecType = 0
	CodecType_CODEC_RC5X_20   CodecType = 1
	CodecType_CODEC_RC5_SZ    CodecType = 2
	CodecType_CODEC_JVC       CodecType = 3
	CodecType_CODEC_SONY12    CodecType = 4
	CodecType_CODEC_SONY15    CodecType = 5
	CodecType_CODEC_SONY20    CodecType = 6
	CodecType_CODEC_NEC16     CodecType = 7
	CodecType_CODEC_NEC32     CodecType = 8
	CodecType_CODEC_NECX      CodecType = 9
	CodecType_CODEC_SANYO     CodecType = 10
	CodecType_CODEC_RC6_0     CodecType = 11
	CodecType_CODEC_RC6_6A_20 CodecType = 12
	CodecType_CODEC_RC6_6A_24 CodecType = 13
	CodecType_CODEC_RC6_6A_32 CodecType = 14
	CodecType_CODEC_RC6_MCE   CodecType = 15
	CodecType_CODEC_SHARP     CodecType = 16
	CodecType_CODEC_APPLETV   CodecType = 17
	CodecType_CODEC_PANASONIC CodecType = 18
)

var CodecType_name = map[int32]string{
	0:  "CODEC_RC5",
	1:  "CODEC_RC5X_20",
	2:  "CODEC_RC5_SZ",
	3:  "CODEC_JVC",
	4:  "CODEC_SONY12",
	5:  "CODEC_SONY15",
	6:  "CODEC_SONY20",
	7:  "CODEC_NEC16",
	8:  "CODEC_NEC32",
	9:  "CODEC_NECX",
	10: "CODEC_SANYO",
	11: "CODEC_RC6_0",
	12: "CODEC_RC6_6A_20",
	13: "CODEC_RC6_6A_24",
	14: "CODEC_RC6_6A_32",
	15: "CODEC_RC6_MCE",
	16: "CODEC_SHARP",
	17: "CODEC_APPLETV",
	18: "CODEC_PANASONIC",
}
var CodecType_value = map[string]int32{
	"CODEC_RC5":       0,
	"CODEC_RC5X_20":   1,
	"CODEC_RC5_SZ":    2,
	"CODEC_JVC":       3,
	"CODEC_SONY12":    4,
	"CODEC_SONY15":    5,
	"CODEC_SONY20":    6,
	"CODEC_NEC16":     7,
	"CODEC_NEC32":     8,
	"CODEC_NECX":      9,
	"CODEC_SANYO":     10,
	"CODEC_RC6_0":     11,
	"CODEC_RC6_6A_20": 12,
	"CODEC_RC6_6A_24": 13,
	"CODEC_RC6_6A_32": 14,
	"CODEC_RC6_MCE":   15,
	"CODEC_SHARP":     16,
	"CODEC_APPLETV":   17,
	"CODEC_PANASONIC": 18,
}

func (x CodecType) String() string {
	return proto.EnumName(CodecType_name, int32(x))
}
func (CodecType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type EmptyRequest struct {
}

func (m *EmptyRequest) Reset()                    { *m = EmptyRequest{} }
func (m *EmptyRequest) String() string            { return proto.CompactTextString(m) }
func (*EmptyRequest) ProtoMessage()               {}
func (*EmptyRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type EmptyReply struct {
}

func (m *EmptyReply) Reset()                    { *m = EmptyReply{} }
func (m *EmptyReply) String() string            { return proto.CompactTextString(m) }
func (*EmptyReply) ProtoMessage()               {}
func (*EmptyReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Point struct {
	X float32 `protobuf:"fixed32,1,opt,name=x" json:"x,omitempty"`
	Y float32 `protobuf:"fixed32,2,opt,name=y" json:"y,omitempty"`
}

func (m *Point) Reset()                    { *m = Point{} }
func (m *Point) String() string            { return proto.CompactTextString(m) }
func (*Point) ProtoMessage()               {}
func (*Point) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Point) GetX() float32 {
	if m != nil {
		return m.X
	}
	return 0
}

func (m *Point) GetY() float32 {
	if m != nil {
		return m.Y
	}
	return 0
}

type InputEvent struct {
	Ts         *google_protobuf.Duration `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	DeviceType InputDeviceType           `protobuf:"varint,2,opt,name=device_type,json=deviceType,enum=mutablelogic.InputDeviceType" json:"device_type,omitempty"`
	EventType  InputEventType            `protobuf:"varint,3,opt,name=event_type,json=eventType,enum=mutablelogic.InputEventType" json:"event_type,omitempty"`
	Scancode   uint32                    `protobuf:"varint,5,opt,name=scancode" json:"scancode,omitempty"`
	Device     uint32                    `protobuf:"varint,6,opt,name=device" json:"device,omitempty"`
	Position   *Point                    `protobuf:"bytes,7,opt,name=position" json:"position,omitempty"`
	Relative   *Point                    `protobuf:"bytes,8,opt,name=relative" json:"relative,omitempty"`
	Slot       uint32                    `protobuf:"varint,9,opt,name=slot" json:"slot,omitempty"`
}

func (m *InputEvent) Reset()                    { *m = InputEvent{} }
func (m *InputEvent) String() string            { return proto.CompactTextString(m) }
func (*InputEvent) ProtoMessage()               {}
func (*InputEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *InputEvent) GetTs() *google_protobuf.Duration {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *InputEvent) GetDeviceType() InputDeviceType {
	if m != nil {
		return m.DeviceType
	}
	return InputDeviceType_INPUT_TYPE_NONE
}

func (m *InputEvent) GetEventType() InputEventType {
	if m != nil {
		return m.EventType
	}
	return InputEventType_INPUT_EVENT_NONE
}

func (m *InputEvent) GetScancode() uint32 {
	if m != nil {
		return m.Scancode
	}
	return 0
}

func (m *InputEvent) GetDevice() uint32 {
	if m != nil {
		return m.Device
	}
	return 0
}

func (m *InputEvent) GetPosition() *Point {
	if m != nil {
		return m.Position
	}
	return nil
}

func (m *InputEvent) GetRelative() *Point {
	if m != nil {
		return m.Relative
	}
	return nil
}

func (m *InputEvent) GetSlot() uint32 {
	if m != nil {
		return m.Slot
	}
	return 0
}

type ReceiveReply struct {
	Event *InputEvent `protobuf:"bytes,1,opt,name=event" json:"event,omitempty"`
	Codec CodecType   `protobuf:"varint,2,opt,name=codec,enum=mutablelogic.CodecType" json:"codec,omitempty"`
}

func (m *ReceiveReply) Reset()                    { *m = ReceiveReply{} }
func (m *ReceiveReply) String() string            { return proto.CompactTextString(m) }
func (*ReceiveReply) ProtoMessage()               {}
func (*ReceiveReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ReceiveReply) GetEvent() *InputEvent {
	if m != nil {
		return m.Event
	}
	return nil
}

func (m *ReceiveReply) GetCodec() CodecType {
	if m != nil {
		return m.Codec
	}
	return CodecType_CODEC_RC5
}

type SendScancodeRequest struct {
	Codec    CodecType `protobuf:"varint,1,opt,name=codec,enum=mutablelogic.CodecType" json:"codec,omitempty"`
	Device   uint32    `protobuf:"varint,2,opt,name=device" json:"device,omitempty"`
	Scancode uint32    `protobuf:"varint,3,opt,name=scancode" json:"scancode,omitempty"`
	Repeats  uint32    `protobuf:"varint,4,opt,name=repeats" json:"repeats,omitempty"`
}

func (m *SendScancodeRequest) Reset()                    { *m = SendScancodeRequest{} }
func (m *SendScancodeRequest) String() string            { return proto.CompactTextString(m) }
func (*SendScancodeRequest) ProtoMessage()               {}
func (*SendScancodeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *SendScancodeRequest) GetCodec() CodecType {
	if m != nil {
		return m.Codec
	}
	return CodecType_CODEC_RC5
}

func (m *SendScancodeRequest) GetDevice() uint32 {
	if m != nil {
		return m.Device
	}
	return 0
}

func (m *SendScancodeRequest) GetScancode() uint32 {
	if m != nil {
		return m.Scancode
	}
	return 0
}

func (m *SendScancodeRequest) GetRepeats() uint32 {
	if m != nil {
		return m.Repeats
	}
	return 0
}

type CodecsReply struct {
	Codec []CodecType `protobuf:"varint,1,rep,packed,name=codec,enum=mutablelogic.CodecType" json:"codec,omitempty"`
}

func (m *CodecsReply) Reset()                    { *m = CodecsReply{} }
func (m *CodecsReply) String() string            { return proto.CompactTextString(m) }
func (*CodecsReply) ProtoMessage()               {}
func (*CodecsReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *CodecsReply) GetCodec() []CodecType {
	if m != nil {
		return m.Codec
	}
	return nil
}

func init() {
	proto.RegisterType((*EmptyRequest)(nil), "mutablelogic.EmptyRequest")
	proto.RegisterType((*EmptyReply)(nil), "mutablelogic.EmptyReply")
	proto.RegisterType((*Point)(nil), "mutablelogic.Point")
	proto.RegisterType((*InputEvent)(nil), "mutablelogic.InputEvent")
	proto.RegisterType((*ReceiveReply)(nil), "mutablelogic.ReceiveReply")
	proto.RegisterType((*SendScancodeRequest)(nil), "mutablelogic.SendScancodeRequest")
	proto.RegisterType((*CodecsReply)(nil), "mutablelogic.CodecsReply")
	proto.RegisterEnum("mutablelogic.InputDeviceType", InputDeviceType_name, InputDeviceType_value)
	proto.RegisterEnum("mutablelogic.InputEventType", InputEventType_name, InputEventType_value)
	proto.RegisterEnum("mutablelogic.CodecType", CodecType_name, CodecType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Remotes service

type RemotesClient interface {
	// Return array of codecs supported
	Codecs(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*CodecsReply, error)
	// Send a remote scancode
	SendScancode(ctx context.Context, in *SendScancodeRequest, opts ...grpc.CallOption) (*EmptyReply, error)
	// Receive remote events
	Receive(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (Remotes_ReceiveClient, error)
}

type remotesClient struct {
	cc *grpc.ClientConn
}

func NewRemotesClient(cc *grpc.ClientConn) RemotesClient {
	return &remotesClient{cc}
}

func (c *remotesClient) Codecs(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*CodecsReply, error) {
	out := new(CodecsReply)
	err := grpc.Invoke(ctx, "/mutablelogic.Remotes/Codecs", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remotesClient) SendScancode(ctx context.Context, in *SendScancodeRequest, opts ...grpc.CallOption) (*EmptyReply, error) {
	out := new(EmptyReply)
	err := grpc.Invoke(ctx, "/mutablelogic.Remotes/SendScancode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remotesClient) Receive(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (Remotes_ReceiveClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Remotes_serviceDesc.Streams[0], c.cc, "/mutablelogic.Remotes/Receive", opts...)
	if err != nil {
		return nil, err
	}
	x := &remotesReceiveClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Remotes_ReceiveClient interface {
	Recv() (*ReceiveReply, error)
	grpc.ClientStream
}

type remotesReceiveClient struct {
	grpc.ClientStream
}

func (x *remotesReceiveClient) Recv() (*ReceiveReply, error) {
	m := new(ReceiveReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Remotes service

type RemotesServer interface {
	// Return array of codecs supported
	Codecs(context.Context, *EmptyRequest) (*CodecsReply, error)
	// Send a remote scancode
	SendScancode(context.Context, *SendScancodeRequest) (*EmptyReply, error)
	// Receive remote events
	Receive(*EmptyRequest, Remotes_ReceiveServer) error
}

func RegisterRemotesServer(s *grpc.Server, srv RemotesServer) {
	s.RegisterService(&_Remotes_serviceDesc, srv)
}

func _Remotes_Codecs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemotesServer).Codecs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mutablelogic.Remotes/Codecs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemotesServer).Codecs(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Remotes_SendScancode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendScancodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RemotesServer).SendScancode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mutablelogic.Remotes/SendScancode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RemotesServer).SendScancode(ctx, req.(*SendScancodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Remotes_Receive_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EmptyRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RemotesServer).Receive(m, &remotesReceiveServer{stream})
}

type Remotes_ReceiveServer interface {
	Send(*ReceiveReply) error
	grpc.ServerStream
}

type remotesReceiveServer struct {
	grpc.ServerStream
}

func (x *remotesReceiveServer) Send(m *ReceiveReply) error {
	return x.ServerStream.SendMsg(m)
}

var _Remotes_serviceDesc = grpc.ServiceDesc{
	ServiceName: "mutablelogic.Remotes",
	HandlerType: (*RemotesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Codecs",
			Handler:    _Remotes_Codecs_Handler,
		},
		{
			MethodName: "SendScancode",
			Handler:    _Remotes_SendScancode_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Receive",
			Handler:       _Remotes_Receive_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "remotes/remotes.proto",
}

func init() { proto.RegisterFile("remotes/remotes.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 826 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0x0e, 0x29, 0xeb, 0x6f, 0x44, 0xcb, 0xeb, 0x75, 0x1c, 0xd3, 0x6a, 0x52, 0xa4, 0xea, 0x25,
	0x35, 0x50, 0x59, 0xa1, 0x1b, 0x5d, 0x5a, 0xb4, 0xa0, 0xa9, 0x05, 0xa2, 0xc8, 0x26, 0x89, 0x25,
	0x6d, 0x44, 0xb9, 0x10, 0xb2, 0xb4, 0x35, 0x04, 0x48, 0x22, 0x2b, 0x52, 0x42, 0xf4, 0x0a, 0x3d,
	0xf6, 0x0d, 0xfa, 0x5e, 0x45, 0x2f, 0x7d, 0x91, 0x82, 0x4b, 0x8a, 0x3f, 0x92, 0xea, 0xe6, 0xc4,
	0x9d, 0xef, 0x9b, 0x99, 0x6f, 0x66, 0xb8, 0x3f, 0x70, 0xba, 0x60, 0x33, 0x37, 0x60, 0xfe, 0x65,
	0xfc, 0x6d, 0x79, 0x0b, 0x37, 0x70, 0xb1, 0x34, 0x5b, 0x06, 0xc3, 0x87, 0x29, 0x9b, 0xba, 0x8f,
	0x93, 0x51, 0xe3, 0xeb, 0x47, 0xd7, 0x7d, 0x9c, 0xb2, 0x4b, 0xce, 0x3d, 0x2c, 0x7f, 0xbd, 0x1c,
	0x2f, 0x17, 0xc3, 0x60, 0xe2, 0xce, 0x23, 0xef, 0x66, 0x1d, 0x24, 0x32, 0xf3, 0x82, 0x35, 0x65,
	0xbf, 0x2d, 0x99, 0x1f, 0x34, 0x25, 0x80, 0xd8, 0xf6, 0xa6, 0xeb, 0xe6, 0xb7, 0x50, 0x34, 0xdd,
	0xc9, 0x3c, 0xc0, 0x12, 0x08, 0x9f, 0x65, 0xe1, 0xb5, 0xf0, 0x46, 0xa4, 0xc2, 0xe7, 0xd0, 0x5a,
	0xcb, 0x62, 0x64, 0xad, 0x9b, 0x7f, 0x8b, 0x00, 0xbd, 0xb9, 0xb7, 0x0c, 0xc8, 0x8a, 0xcd, 0x03,
	0xfc, 0x1d, 0x88, 0x81, 0xcf, 0x7d, 0x6b, 0xca, 0x79, 0x2b, 0x92, 0x6f, 0x6d, 0xe4, 0x5b, 0xdd,
	0x58, 0x9e, 0x8a, 0x81, 0x8f, 0x7f, 0x86, 0xda, 0x98, 0xad, 0x26, 0x23, 0xe6, 0x04, 0x6b, 0x8f,
	0xf1, 0x8c, 0x75, 0xe5, 0x55, 0x2b, 0xdb, 0x40, 0x8b, 0x67, 0xee, 0x72, 0x2f, 0x7b, 0xed, 0x31,
	0x0a, 0xe3, 0x64, 0x8d, 0x7f, 0x04, 0x60, 0xa1, 0x66, 0x14, 0x5e, 0xe0, 0xe1, 0x2f, 0xf7, 0x84,
	0xf3, 0xc2, 0x78, 0x74, 0x95, 0x6d, 0x96, 0xb8, 0x01, 0x15, 0x7f, 0x34, 0x9c, 0x8f, 0xdc, 0x31,
	0x93, 0x8b, 0xaf, 0x85, 0x37, 0x87, 0x34, 0xb1, 0xf1, 0x0b, 0x28, 0x45, 0x32, 0x72, 0x89, 0x33,
	0xb1, 0x85, 0x2f, 0xa1, 0xe2, 0xb9, 0xfe, 0x24, 0x6c, 0x40, 0x2e, 0xf3, 0x0e, 0x4f, 0xf2, 0x72,
	0x7c, 0x5a, 0x34, 0x71, 0x0a, 0x03, 0x16, 0x6c, 0x3a, 0x0c, 0x26, 0x2b, 0x26, 0x57, 0x9e, 0x08,
	0xd8, 0x38, 0x61, 0x0c, 0x07, 0xfe, 0xd4, 0x0d, 0xe4, 0x2a, 0xd7, 0xe5, 0xeb, 0xe6, 0x0c, 0x24,
	0xca, 0x46, 0x6c, 0xb2, 0x62, 0xfc, 0xaf, 0xe0, 0x16, 0x14, 0x79, 0x1b, 0xf1, 0x90, 0xe5, 0xff,
	0xea, 0x98, 0x46, 0x6e, 0xf8, 0x7b, 0x28, 0x86, 0x5d, 0x8d, 0xe2, 0x01, 0x9f, 0xe5, 0xfd, 0xb5,
	0x90, 0xe2, 0xc3, 0x89, 0xbc, 0x9a, 0x7f, 0x08, 0x70, 0x62, 0xb1, 0xf9, 0xd8, 0x8a, 0xa7, 0x11,
	0x6f, 0x8d, 0x34, 0x8d, 0xf0, 0x25, 0x69, 0x32, 0x33, 0x14, 0x73, 0x33, 0xcc, 0xce, 0xbd, 0xb0,
	0x35, 0x77, 0x19, 0xca, 0x0b, 0xe6, 0xb1, 0x61, 0xe0, 0xcb, 0x07, 0x9c, 0xda, 0x98, 0xcd, 0x9f,
	0xa0, 0xc6, 0x15, 0xfc, 0x68, 0x04, 0x99, 0x5a, 0x0a, 0xff, 0x5f, 0xcb, 0xc5, 0x9f, 0x02, 0x1c,
	0x6d, 0x6d, 0x24, 0x7c, 0x02, 0x47, 0x3d, 0xdd, 0xbc, 0xb3, 0x1d, 0x7b, 0x60, 0x12, 0x47, 0x37,
	0x74, 0x82, 0x9e, 0xe1, 0x33, 0x38, 0xc9, 0x80, 0x7d, 0x32, 0xb8, 0x36, 0x54, 0xda, 0x45, 0x02,
	0x7e, 0x0e, 0x28, 0x43, 0xdc, 0x1a, 0x77, 0x16, 0x41, 0x22, 0x6e, 0xc0, 0x8b, 0x0c, 0x6a, 0x1b,
	0x77, 0xda, 0x7b, 0x4b, 0xa3, 0x84, 0xe8, 0xe8, 0x60, 0x2b, 0xd5, 0x07, 0x63, 0x60, 0xd9, 0x3d,
	0xad, 0x8f, 0x2a, 0xf8, 0x14, 0x8e, 0x33, 0x04, 0x25, 0xb7, 0x86, 0x4d, 0x10, 0xba, 0xf8, 0x5d,
	0x84, 0x7a, 0x7e, 0xb7, 0xa6, 0xa2, 0xe4, 0x9e, 0xe8, 0xf6, 0xa6, 0x46, 0x19, 0x9e, 0x67, 0xd1,
	0x3e, 0x19, 0x98, 0x94, 0x58, 0x16, 0x12, 0xd2, 0x72, 0x12, 0x86, 0x92, 0x1b, 0xa2, 0xf2, 0x52,
	0xcf, 0xe1, 0x74, 0x87, 0x33, 0x89, 0x6a, 0xa3, 0x02, 0xfe, 0x0a, 0xce, 0xb2, 0x94, 0x7a, 0x6d,
	0x99, 0x86, 0xd5, 0xb3, 0x7b, 0x46, 0xd8, 0xc6, 0x16, 0x49, 0xc9, 0x4d, 0x42, 0x16, 0xb7, 0x05,
	0xf9, 0x00, 0xa2, 0x62, 0x4a, 0xf8, 0x25, 0xc8, 0x3b, 0xdc, 0xa6, 0x9c, 0x32, 0x7e, 0x05, 0xe7,
	0xbb, 0x91, 0x9b, 0xc4, 0x95, 0x8b, 0x7f, 0x44, 0xa8, 0x26, 0x7f, 0x11, 0x1f, 0x42, 0x55, 0x33,
	0xba, 0x44, 0x73, 0xa8, 0xf6, 0x0e, 0x3d, 0xc3, 0xc7, 0x70, 0x98, 0x98, 0x1f, 0x1d, 0xa5, 0x8d,
	0x04, 0x8c, 0x40, 0x4a, 0x20, 0xc7, 0xfa, 0x84, 0xc4, 0x34, 0xe6, 0xc3, 0xbd, 0x86, 0x0a, 0xa9,
	0x83, 0x65, 0xe8, 0x83, 0xb7, 0x0a, 0x3a, 0xd8, 0x42, 0xde, 0xa1, 0x62, 0x1e, 0x51, 0xda, 0xa8,
	0x84, 0x8f, 0xa0, 0x16, 0x21, 0x3a, 0xd1, 0xde, 0x76, 0x50, 0x39, 0x07, 0x5c, 0x29, 0xa8, 0x82,
	0xeb, 0x00, 0x09, 0xf0, 0x11, 0x55, 0x53, 0x07, 0x4b, 0xd5, 0x07, 0x06, 0x82, 0x14, 0xa0, 0x5a,
	0xc7, 0x69, 0xa3, 0x5a, 0xb8, 0xef, 0x52, 0xa0, 0xa3, 0x86, 0xf5, 0x4b, 0xbb, 0xe0, 0x0f, 0xe8,
	0x70, 0x07, 0xbc, 0x52, 0x50, 0x3d, 0xdb, 0x7c, 0xc7, 0xb9, 0xd5, 0x08, 0x3a, 0xca, 0x68, 0xbe,
	0x57, 0xa9, 0x89, 0x50, 0xea, 0xa3, 0x9a, 0xe6, 0x0d, 0xb1, 0xef, 0xd1, 0x71, 0x9a, 0xcb, 0x54,
	0x75, 0xd5, 0x32, 0xf4, 0x9e, 0x86, 0xb0, 0xf2, 0x97, 0x00, 0x65, 0x1a, 0x3d, 0x1e, 0xf8, 0x17,
	0x28, 0x45, 0x07, 0x0c, 0x37, 0xf2, 0x87, 0x29, 0xfb, 0x3c, 0x34, 0xce, 0xf7, 0x1c, 0xb4, 0xf8,
	0x48, 0xf6, 0x41, 0xca, 0xde, 0x1a, 0xf8, 0x9b, 0xbc, 0xeb, 0x9e, 0x1b, 0xa5, 0x21, 0xef, 0x55,
	0x0a, 0x93, 0x69, 0x61, 0x61, 0xfc, 0xca, 0x7b, 0xb2, 0x9c, 0x2d, 0x2e, 0x7b, 0x4b, 0xb6, 0x85,
	0xeb, 0xea, 0xa7, 0x72, 0xfc, 0x34, 0x3e, 0x94, 0xf8, 0x03, 0x74, 0xf5, 0x6f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x46, 0x15, 0xd9, 0xec, 0x34, 0x07, 0x00, 0x00,
}
