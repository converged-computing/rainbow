// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.20.3
// source: api.proto

package v1

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

// RainbowSchedulerClient is the client API for RainbowScheduler service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RainbowSchedulerClient interface {
	// Register cluster - request to register a new cluster
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	// Job Submission - request for submitting a job to a named cluster
	SubmitJob(ctx context.Context, in *SubmitJobRequest, opts ...grpc.CallOption) (*SubmitJobResponse, error)
	// Request Job - ask the rainbow scheduler for up to max jobs
	RequestJobs(ctx context.Context, in *RequestJobsRequest, opts ...grpc.CallOption) (*RequestJobsResponse, error)
	// TESTING ENDPOINTS
	// Serial checks the connectivity and response time of the service.
	Serial(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// Stream continuously sends and receives response messages.
	// It is useful for scenarios where constant data flow is required.
	Stream(ctx context.Context, opts ...grpc.CallOption) (RainbowScheduler_StreamClient, error)
}

type rainbowSchedulerClient struct {
	cc grpc.ClientConnInterface
}

func NewRainbowSchedulerClient(cc grpc.ClientConnInterface) RainbowSchedulerClient {
	return &rainbowSchedulerClient{cc}
}

func (c *rainbowSchedulerClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/convergedcomputing.org.grpc.v1.RainbowScheduler/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rainbowSchedulerClient) SubmitJob(ctx context.Context, in *SubmitJobRequest, opts ...grpc.CallOption) (*SubmitJobResponse, error) {
	out := new(SubmitJobResponse)
	err := c.cc.Invoke(ctx, "/convergedcomputing.org.grpc.v1.RainbowScheduler/SubmitJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rainbowSchedulerClient) RequestJobs(ctx context.Context, in *RequestJobsRequest, opts ...grpc.CallOption) (*RequestJobsResponse, error) {
	out := new(RequestJobsResponse)
	err := c.cc.Invoke(ctx, "/convergedcomputing.org.grpc.v1.RainbowScheduler/RequestJobs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rainbowSchedulerClient) Serial(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/convergedcomputing.org.grpc.v1.RainbowScheduler/Serial", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rainbowSchedulerClient) Stream(ctx context.Context, opts ...grpc.CallOption) (RainbowScheduler_StreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &RainbowScheduler_ServiceDesc.Streams[0], "/convergedcomputing.org.grpc.v1.RainbowScheduler/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &rainbowSchedulerStreamClient{stream}
	return x, nil
}

type RainbowScheduler_StreamClient interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type rainbowSchedulerStreamClient struct {
	grpc.ClientStream
}

func (x *rainbowSchedulerStreamClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rainbowSchedulerStreamClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RainbowSchedulerServer is the server API for RainbowScheduler service.
// All implementations must embed UnimplementedRainbowSchedulerServer
// for forward compatibility
type RainbowSchedulerServer interface {
	// Register cluster - request to register a new cluster
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	// Job Submission - request for submitting a job to a named cluster
	SubmitJob(context.Context, *SubmitJobRequest) (*SubmitJobResponse, error)
	// Request Job - ask the rainbow scheduler for up to max jobs
	RequestJobs(context.Context, *RequestJobsRequest) (*RequestJobsResponse, error)
	// TESTING ENDPOINTS
	// Serial checks the connectivity and response time of the service.
	Serial(context.Context, *Request) (*Response, error)
	// Stream continuously sends and receives response messages.
	// It is useful for scenarios where constant data flow is required.
	Stream(RainbowScheduler_StreamServer) error
	mustEmbedUnimplementedRainbowSchedulerServer()
}

// UnimplementedRainbowSchedulerServer must be embedded to have forward compatible implementations.
type UnimplementedRainbowSchedulerServer struct {
}

func (UnimplementedRainbowSchedulerServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedRainbowSchedulerServer) SubmitJob(context.Context, *SubmitJobRequest) (*SubmitJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitJob not implemented")
}
func (UnimplementedRainbowSchedulerServer) RequestJobs(context.Context, *RequestJobsRequest) (*RequestJobsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestJobs not implemented")
}
func (UnimplementedRainbowSchedulerServer) Serial(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Serial not implemented")
}
func (UnimplementedRainbowSchedulerServer) Stream(RainbowScheduler_StreamServer) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedRainbowSchedulerServer) mustEmbedUnimplementedRainbowSchedulerServer() {}

// UnsafeRainbowSchedulerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RainbowSchedulerServer will
// result in compilation errors.
type UnsafeRainbowSchedulerServer interface {
	mustEmbedUnimplementedRainbowSchedulerServer()
}

func RegisterRainbowSchedulerServer(s grpc.ServiceRegistrar, srv RainbowSchedulerServer) {
	s.RegisterService(&RainbowScheduler_ServiceDesc, srv)
}

func _RainbowScheduler_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RainbowSchedulerServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/convergedcomputing.org.grpc.v1.RainbowScheduler/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RainbowSchedulerServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RainbowScheduler_SubmitJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RainbowSchedulerServer).SubmitJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/convergedcomputing.org.grpc.v1.RainbowScheduler/SubmitJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RainbowSchedulerServer).SubmitJob(ctx, req.(*SubmitJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RainbowScheduler_RequestJobs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestJobsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RainbowSchedulerServer).RequestJobs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/convergedcomputing.org.grpc.v1.RainbowScheduler/RequestJobs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RainbowSchedulerServer).RequestJobs(ctx, req.(*RequestJobsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RainbowScheduler_Serial_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RainbowSchedulerServer).Serial(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/convergedcomputing.org.grpc.v1.RainbowScheduler/Serial",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RainbowSchedulerServer).Serial(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _RainbowScheduler_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RainbowSchedulerServer).Stream(&rainbowSchedulerStreamServer{stream})
}

type RainbowScheduler_StreamServer interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type rainbowSchedulerStreamServer struct {
	grpc.ServerStream
}

func (x *rainbowSchedulerStreamServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rainbowSchedulerStreamServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RainbowScheduler_ServiceDesc is the grpc.ServiceDesc for RainbowScheduler service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RainbowScheduler_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "convergedcomputing.org.grpc.v1.RainbowScheduler",
	HandlerType: (*RainbowSchedulerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _RainbowScheduler_Register_Handler,
		},
		{
			MethodName: "SubmitJob",
			Handler:    _RainbowScheduler_SubmitJob_Handler,
		},
		{
			MethodName: "RequestJobs",
			Handler:    _RainbowScheduler_RequestJobs_Handler,
		},
		{
			MethodName: "Serial",
			Handler:    _RainbowScheduler_Serial_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _RainbowScheduler_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "api.proto",
}
