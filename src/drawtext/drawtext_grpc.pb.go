// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package drawtext

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DrawTextClient is the client API for DrawText service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DrawTextClient interface {
	Draw(ctx context.Context, in *DrawRequest, opts ...grpc.CallOption) (*DrawReply, error)
	GetFontNames(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TextReply, error)
}

type drawTextClient struct {
	cc grpc.ClientConnInterface
}

func NewDrawTextClient(cc grpc.ClientConnInterface) DrawTextClient {
	return &drawTextClient{cc}
}

func (c *drawTextClient) Draw(ctx context.Context, in *DrawRequest, opts ...grpc.CallOption) (*DrawReply, error) {
	out := new(DrawReply)
	err := c.cc.Invoke(ctx, "/DrawText/Draw", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *drawTextClient) GetFontNames(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*TextReply, error) {
	out := new(TextReply)
	err := c.cc.Invoke(ctx, "/DrawText/GetFontNames", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DrawTextServer is the server API for DrawText service.
// All implementations must embed UnimplementedDrawTextServer
// for forward compatibility
type DrawTextServer interface {
	Draw(context.Context, *DrawRequest) (*DrawReply, error)
	GetFontNames(context.Context, *emptypb.Empty) (*TextReply, error)
	mustEmbedUnimplementedDrawTextServer()
}

// UnimplementedDrawTextServer must be embedded to have forward compatible implementations.
type UnimplementedDrawTextServer struct {
}

func (UnimplementedDrawTextServer) Draw(context.Context, *DrawRequest) (*DrawReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Draw not implemented")
}
func (UnimplementedDrawTextServer) GetFontNames(context.Context, *emptypb.Empty) (*TextReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFontNames not implemented")
}
func (UnimplementedDrawTextServer) mustEmbedUnimplementedDrawTextServer() {}

// UnsafeDrawTextServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DrawTextServer will
// result in compilation errors.
type UnsafeDrawTextServer interface {
	mustEmbedUnimplementedDrawTextServer()
}

func RegisterDrawTextServer(s grpc.ServiceRegistrar, srv DrawTextServer) {
	s.RegisterService(&DrawText_ServiceDesc, srv)
}

func _DrawText_Draw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DrawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DrawTextServer).Draw(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DrawText/Draw",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DrawTextServer).Draw(ctx, req.(*DrawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DrawText_GetFontNames_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DrawTextServer).GetFontNames(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DrawText/GetFontNames",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DrawTextServer).GetFontNames(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// DrawText_ServiceDesc is the grpc.ServiceDesc for DrawText service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DrawText_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DrawText",
	HandlerType: (*DrawTextServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Draw",
			Handler:    _DrawText_Draw_Handler,
		},
		{
			MethodName: "GetFontNames",
			Handler:    _DrawText_GetFontNames_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "drawtext/drawtext.proto",
}
