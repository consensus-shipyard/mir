// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package requestreceiver

import (
	context "context"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RequestReceiverClient is the client API for RequestReceiver service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RequestReceiverClient interface {
	Listen(ctx context.Context, opts ...grpc.CallOption) (RequestReceiver_ListenClient, error)
}

type requestReceiverClient struct {
	cc grpc.ClientConnInterface
}

func NewRequestReceiverClient(cc grpc.ClientConnInterface) RequestReceiverClient {
	return &requestReceiverClient{cc}
}

func (c *requestReceiverClient) Listen(ctx context.Context, opts ...grpc.CallOption) (RequestReceiver_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &RequestReceiver_ServiceDesc.Streams[0], "/requestreceiver.RequestReceiver/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &requestReceiverListenClient{stream}
	return x, nil
}

type RequestReceiver_ListenClient interface {
	Send(*requestpb.Request) error
	CloseAndRecv() (*ByeBye, error)
	grpc.ClientStream
}

type requestReceiverListenClient struct {
	grpc.ClientStream
}

func (x *requestReceiverListenClient) Send(m *requestpb.Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *requestReceiverListenClient) CloseAndRecv() (*ByeBye, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ByeBye)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RequestReceiverServer is the server API for RequestReceiver service.
// All implementations must embed UnimplementedRequestReceiverServer
// for forward compatibility
type RequestReceiverServer interface {
	Listen(RequestReceiver_ListenServer) error
	mustEmbedUnimplementedRequestReceiverServer()
}

// UnimplementedRequestReceiverServer must be embedded to have forward compatible implementations.
type UnimplementedRequestReceiverServer struct {
}

func (UnimplementedRequestReceiverServer) Listen(RequestReceiver_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (UnimplementedRequestReceiverServer) mustEmbedUnimplementedRequestReceiverServer() {}

// UnsafeRequestReceiverServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RequestReceiverServer will
// result in compilation errors.
type UnsafeRequestReceiverServer interface {
	mustEmbedUnimplementedRequestReceiverServer()
}

func RegisterRequestReceiverServer(s grpc.ServiceRegistrar, srv RequestReceiverServer) {
	s.RegisterService(&RequestReceiver_ServiceDesc, srv)
}

func _RequestReceiver_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RequestReceiverServer).Listen(&requestReceiverListenServer{stream})
}

type RequestReceiver_ListenServer interface {
	SendAndClose(*ByeBye) error
	Recv() (*requestpb.Request, error)
	grpc.ServerStream
}

type requestReceiverListenServer struct {
	grpc.ServerStream
}

func (x *requestReceiverListenServer) SendAndClose(m *ByeBye) error {
	return x.ServerStream.SendMsg(m)
}

func (x *requestReceiverListenServer) Recv() (*requestpb.Request, error) {
	m := new(requestpb.Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RequestReceiver_ServiceDesc is the grpc.ServiceDesc for RequestReceiver service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RequestReceiver_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "requestreceiver.RequestReceiver",
	HandlerType: (*RequestReceiverServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _RequestReceiver_Listen_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "requestreceiver/requestreceiver.proto",
}
