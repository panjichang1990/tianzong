// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package service

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

// ChildClient is the client API for Child service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChildClient interface {
	Do(ctx context.Context, in *DoReq, opts ...grpc.CallOption) (*DoRep, error)
	ReloadChannel(ctx context.Context, in *ReloadChannelReq, opts ...grpc.CallOption) (*ReloadRep, error)
	Publish(ctx context.Context, in *PublishInfo, opts ...grpc.CallOption) (*PublishRep, error)
}

type childClient struct {
	cc grpc.ClientConnInterface
}

func NewChildClient(cc grpc.ClientConnInterface) ChildClient {
	return &childClient{cc}
}

func (c *childClient) Do(ctx context.Context, in *DoReq, opts ...grpc.CallOption) (*DoRep, error) {
	out := new(DoRep)
	err := c.cc.Invoke(ctx, "/Child/Do", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *childClient) ReloadChannel(ctx context.Context, in *ReloadChannelReq, opts ...grpc.CallOption) (*ReloadRep, error) {
	out := new(ReloadRep)
	err := c.cc.Invoke(ctx, "/Child/ReloadChannel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *childClient) Publish(ctx context.Context, in *PublishInfo, opts ...grpc.CallOption) (*PublishRep, error) {
	out := new(PublishRep)
	err := c.cc.Invoke(ctx, "/Child/Publish", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChildServer is the server API for Child service.
// All implementations must embed UnimplementedChildServer
// for forward compatibility
type ChildServer interface {
	Do(context.Context, *DoReq) (*DoRep, error)
	ReloadChannel(context.Context, *ReloadChannelReq) (*ReloadRep, error)
	Publish(context.Context, *PublishInfo) (*PublishRep, error)
	mustEmbedUnimplementedChildServer()
}

// UnimplementedChildServer must be embedded to have forward compatible implementations.
type UnimplementedChildServer struct {
}

func (UnimplementedChildServer) Do(context.Context, *DoReq) (*DoRep, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Do not implemented")
}
func (UnimplementedChildServer) ReloadChannel(context.Context, *ReloadChannelReq) (*ReloadRep, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReloadChannel not implemented")
}
func (UnimplementedChildServer) Publish(context.Context, *PublishInfo) (*PublishRep, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Publish not implemented")
}
func (UnimplementedChildServer) mustEmbedUnimplementedChildServer() {}

// UnsafeChildServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChildServer will
// result in compilation errors.
type UnsafeChildServer interface {
	mustEmbedUnimplementedChildServer()
}

func RegisterChildServer(s grpc.ServiceRegistrar, srv ChildServer) {
	s.RegisterService(&Child_ServiceDesc, srv)
}

func _Child_Do_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChildServer).Do(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Child/Do",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChildServer).Do(ctx, req.(*DoReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Child_ReloadChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReloadChannelReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChildServer).ReloadChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Child/ReloadChannel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChildServer).ReloadChannel(ctx, req.(*ReloadChannelReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Child_Publish_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChildServer).Publish(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Child/Publish",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChildServer).Publish(ctx, req.(*PublishInfo))
	}
	return interceptor(ctx, in, info, handler)
}

// Child_ServiceDesc is the grpc.ServiceDesc for Child service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Child_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Child",
	HandlerType: (*ChildServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Do",
			Handler:    _Child_Do_Handler,
		},
		{
			MethodName: "ReloadChannel",
			Handler:    _Child_ReloadChannel_Handler,
		},
		{
			MethodName: "Publish",
			Handler:    _Child_Publish_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "child.proto",
}
