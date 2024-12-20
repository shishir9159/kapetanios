// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.1
// source: proto/certs-renewal.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Renewal_ClusterHealthChecking_FullMethodName = "/Renewal/ClusterHealthChecking"
	Renewal_BackupUpdate_FullMethodName          = "/Renewal/BackupUpdate"
	Renewal_RenewalUpdate_FullMethodName         = "/Renewal/RenewalUpdate"
	Renewal_RestartUpdate_FullMethodName         = "/Renewal/RestartUpdate"
)

// RenewalClient is the client API for Renewal service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RenewalClient interface {
	ClusterHealthChecking(ctx context.Context, in *PrerequisitesRenewal, opts ...grpc.CallOption) (*RenewalResponse, error)
	BackupUpdate(ctx context.Context, in *BackupStatus, opts ...grpc.CallOption) (*RenewalResponse, error)
	RenewalUpdate(ctx context.Context, in *RenewalStatus, opts ...grpc.CallOption) (*RenewalResponse, error)
	RestartUpdate(ctx context.Context, in *RestartStatus, opts ...grpc.CallOption) (*RenewalFinalizer, error)
}

type renewalClient struct {
	cc grpc.ClientConnInterface
}

func NewRenewalClient(cc grpc.ClientConnInterface) RenewalClient {
	return &renewalClient{cc}
}

func (c *renewalClient) ClusterHealthChecking(ctx context.Context, in *PrerequisitesRenewal, opts ...grpc.CallOption) (*RenewalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RenewalResponse)
	err := c.cc.Invoke(ctx, Renewal_ClusterHealthChecking_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *renewalClient) BackupUpdate(ctx context.Context, in *BackupStatus, opts ...grpc.CallOption) (*RenewalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RenewalResponse)
	err := c.cc.Invoke(ctx, Renewal_BackupUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *renewalClient) RenewalUpdate(ctx context.Context, in *RenewalStatus, opts ...grpc.CallOption) (*RenewalResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RenewalResponse)
	err := c.cc.Invoke(ctx, Renewal_RenewalUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *renewalClient) RestartUpdate(ctx context.Context, in *RestartStatus, opts ...grpc.CallOption) (*RenewalFinalizer, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RenewalFinalizer)
	err := c.cc.Invoke(ctx, Renewal_RestartUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RenewalServer is the server API for Renewal service.
// All implementations must embed UnimplementedRenewalServer
// for forward compatibility.
type RenewalServer interface {
	ClusterHealthChecking(context.Context, *PrerequisitesRenewal) (*RenewalResponse, error)
	BackupUpdate(context.Context, *BackupStatus) (*RenewalResponse, error)
	RenewalUpdate(context.Context, *RenewalStatus) (*RenewalResponse, error)
	RestartUpdate(context.Context, *RestartStatus) (*RenewalFinalizer, error)
	mustEmbedUnimplementedRenewalServer()
}

// UnimplementedRenewalServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRenewalServer struct{}

func (UnimplementedRenewalServer) ClusterHealthChecking(context.Context, *PrerequisitesRenewal) (*RenewalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClusterHealthChecking not implemented")
}
func (UnimplementedRenewalServer) BackupUpdate(context.Context, *BackupStatus) (*RenewalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BackupUpdate not implemented")
}
func (UnimplementedRenewalServer) RenewalUpdate(context.Context, *RenewalStatus) (*RenewalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RenewalUpdate not implemented")
}
func (UnimplementedRenewalServer) RestartUpdate(context.Context, *RestartStatus) (*RenewalFinalizer, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RestartUpdate not implemented")
}
func (UnimplementedRenewalServer) mustEmbedUnimplementedRenewalServer() {}
func (UnimplementedRenewalServer) testEmbeddedByValue()                 {}

// UnsafeRenewalServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RenewalServer will
// result in compilation errors.
type UnsafeRenewalServer interface {
	mustEmbedUnimplementedRenewalServer()
}

func RegisterRenewalServer(s grpc.ServiceRegistrar, srv RenewalServer) {
	// If the following call pancis, it indicates UnimplementedRenewalServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Renewal_ServiceDesc, srv)
}

func _Renewal_ClusterHealthChecking_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PrerequisitesRenewal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RenewalServer).ClusterHealthChecking(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Renewal_ClusterHealthChecking_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RenewalServer).ClusterHealthChecking(ctx, req.(*PrerequisitesRenewal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Renewal_BackupUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BackupStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RenewalServer).BackupUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Renewal_BackupUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RenewalServer).BackupUpdate(ctx, req.(*BackupStatus))
	}
	return interceptor(ctx, in, info, handler)
}

func _Renewal_RenewalUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenewalStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RenewalServer).RenewalUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Renewal_RenewalUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RenewalServer).RenewalUpdate(ctx, req.(*RenewalStatus))
	}
	return interceptor(ctx, in, info, handler)
}

func _Renewal_RestartUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RenewalServer).RestartUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Renewal_RestartUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RenewalServer).RestartUpdate(ctx, req.(*RestartStatus))
	}
	return interceptor(ctx, in, info, handler)
}

// Renewal_ServiceDesc is the grpc.ServiceDesc for Renewal service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Renewal_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Renewal",
	HandlerType: (*RenewalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ClusterHealthChecking",
			Handler:    _Renewal_ClusterHealthChecking_Handler,
		},
		{
			MethodName: "BackupUpdate",
			Handler:    _Renewal_BackupUpdate_Handler,
		},
		{
			MethodName: "RenewalUpdate",
			Handler:    _Renewal_RenewalUpdate_Handler,
		},
		{
			MethodName: "RestartUpdate",
			Handler:    _Renewal_RestartUpdate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/certs-renewal.proto",
}
