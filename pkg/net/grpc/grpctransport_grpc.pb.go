// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GrpcTransportClient is the client API for GrpcTransport service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GrpcTransportClient interface {
	Listen(ctx context.Context, opts ...grpc.CallOption) (GrpcTransport_ListenClient, error)
}

type grpcTransportClient struct {
	cc grpc.ClientConnInterface
}

func NewGrpcTransportClient(cc grpc.ClientConnInterface) GrpcTransportClient {
	return &grpcTransportClient{cc}
}

func (c *grpcTransportClient) Listen(ctx context.Context, opts ...grpc.CallOption) (GrpcTransport_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &GrpcTransport_ServiceDesc.Streams[0], "/grpctransport.GrpcTransport/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &grpcTransportListenClient{stream}
	return x, nil
}

type GrpcTransport_ListenClient interface {
	Send(*GrpcMessage) error
	CloseAndRecv() (*ByeBye, error)
	grpc.ClientStream
}

type grpcTransportListenClient struct {
	grpc.ClientStream
}

func (x *grpcTransportListenClient) Send(m *GrpcMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *grpcTransportListenClient) CloseAndRecv() (*ByeBye, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ByeBye)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GrpcTransportServer is the server API for GrpcTransport service.
// All implementations must embed UnimplementedGrpcTransportServer
// for forward compatibility
type GrpcTransportServer interface {
	Listen(GrpcTransport_ListenServer) error
	mustEmbedUnimplementedGrpcTransportServer()
}

// UnimplementedGrpcTransportServer must be embedded to have forward compatible implementations.
type UnimplementedGrpcTransportServer struct {
}

func (UnimplementedGrpcTransportServer) Listen(GrpcTransport_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (UnimplementedGrpcTransportServer) mustEmbedUnimplementedGrpcTransportServer() {}

// UnsafeGrpcTransportServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GrpcTransportServer will
// result in compilation errors.
type UnsafeGrpcTransportServer interface {
	mustEmbedUnimplementedGrpcTransportServer()
}

func RegisterGrpcTransportServer(s grpc.ServiceRegistrar, srv GrpcTransportServer) {
	s.RegisterService(&GrpcTransport_ServiceDesc, srv)
}

func _GrpcTransport_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GrpcTransportServer).Listen(&grpcTransportListenServer{stream})
}

type GrpcTransport_ListenServer interface {
	SendAndClose(*ByeBye) error
	Recv() (*GrpcMessage, error)
	grpc.ServerStream
}

type grpcTransportListenServer struct {
	grpc.ServerStream
}

func (x *grpcTransportListenServer) SendAndClose(m *ByeBye) error {
	return x.ServerStream.SendMsg(m)
}

func (x *grpcTransportListenServer) Recv() (*GrpcMessage, error) {
	m := new(GrpcMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GrpcTransport_ServiceDesc is the grpc.ServiceDesc for GrpcTransport service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GrpcTransport_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpctransport.GrpcTransport",
	HandlerType: (*GrpcTransportServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _GrpcTransport_Listen_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "net/grpc/grpctransport.proto",
}
