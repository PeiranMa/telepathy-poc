// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0-devel
// 	protoc        v3.12.1
// source: req.proto

package protos

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type TaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskId int32  `protobuf:"varint,1,opt,name=TaskId,proto3" json:"TaskId,omitempty"`
	Body   []byte `protobuf:"bytes,2,opt,name=Body,proto3" json:"Body,omitempty"`
}

func (x *TaskRequest) Reset() {
	*x = TaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_req_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskRequest) ProtoMessage() {}

func (x *TaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_req_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskRequest.ProtoReflect.Descriptor instead.
func (*TaskRequest) Descriptor() ([]byte, []int) {
	return file_req_proto_rawDescGZIP(), []int{0}
}

func (x *TaskRequest) GetTaskId() int32 {
	if x != nil {
		return x.TaskId
	}
	return 0
}

func (x *TaskRequest) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

type TaskResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskId int32  `protobuf:"varint,1,opt,name=TaskId,proto3" json:"TaskId,omitempty"`
	Body   []byte `protobuf:"bytes,2,opt,name=Body,proto3" json:"Body,omitempty"`
}

func (x *TaskResponse) Reset() {
	*x = TaskResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_req_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResponse) ProtoMessage() {}

func (x *TaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_req_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResponse.ProtoReflect.Descriptor instead.
func (*TaskResponse) Descriptor() ([]byte, []int) {
	return file_req_proto_rawDescGZIP(), []int{1}
}

func (x *TaskResponse) GetTaskId() int32 {
	if x != nil {
		return x.TaskId
	}
	return 0
}

func (x *TaskResponse) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

var File_req_proto protoreflect.FileDescriptor

var file_req_proto_rawDesc = []byte{
	0x0a, 0x09, 0x72, 0x65, 0x71, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x22, 0x39, 0x0a, 0x0b, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x54, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x54, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x42, 0x6f,
	0x64, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x22, 0x3a,
	0x0a, 0x0c, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x54, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06,
	0x54, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x32, 0x7d, 0x0a, 0x0d, 0x44, 0x69,
	0x73, 0x70, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72, 0x53, 0x76, 0x72, 0x12, 0x35, 0x0a, 0x08, 0x72,
	0x65, 0x71, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x12, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x66, 0x69, 0x6e, 0x5f, 0x74, 0x61, 0x73, 0x6b, 0x12, 0x13,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x54, 0x61, 0x73,
	0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_req_proto_rawDescOnce sync.Once
	file_req_proto_rawDescData = file_req_proto_rawDesc
)

func file_req_proto_rawDescGZIP() []byte {
	file_req_proto_rawDescOnce.Do(func() {
		file_req_proto_rawDescData = protoimpl.X.CompressGZIP(file_req_proto_rawDescData)
	})
	return file_req_proto_rawDescData
}

var file_req_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_req_proto_goTypes = []interface{}{
	(*TaskRequest)(nil),  // 0: protos.TaskRequest
	(*TaskResponse)(nil), // 1: protos.TaskResponse
}
var file_req_proto_depIdxs = []int32{
	0, // 0: protos.DispatcherSvr.req_task:input_type -> protos.TaskRequest
	0, // 1: protos.DispatcherSvr.fin_task:input_type -> protos.TaskRequest
	1, // 2: protos.DispatcherSvr.req_task:output_type -> protos.TaskResponse
	1, // 3: protos.DispatcherSvr.fin_task:output_type -> protos.TaskResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_req_proto_init() }
func file_req_proto_init() {
	if File_req_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_req_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_req_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_req_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_req_proto_goTypes,
		DependencyIndexes: file_req_proto_depIdxs,
		MessageInfos:      file_req_proto_msgTypes,
	}.Build()
	File_req_proto = out.File
	file_req_proto_rawDesc = nil
	file_req_proto_goTypes = nil
	file_req_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// DispatcherSvrClient is the client API for DispatcherSvr service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DispatcherSvrClient interface {
	ReqTask(ctx context.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
	FinTask(ctx context.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error)
}

type dispatcherSvrClient struct {
	cc grpc.ClientConnInterface
}

func NewDispatcherSvrClient(cc grpc.ClientConnInterface) DispatcherSvrClient {
	return &dispatcherSvrClient{cc}
}

func (c *dispatcherSvrClient) ReqTask(ctx context.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, "/protos.DispatcherSvr/req_task", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dispatcherSvrClient) FinTask(ctx context.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	out := new(TaskResponse)
	err := c.cc.Invoke(ctx, "/protos.DispatcherSvr/fin_task", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DispatcherSvrServer is the server API for DispatcherSvr service.
type DispatcherSvrServer interface {
	ReqTask(context.Context, *TaskRequest) (*TaskResponse, error)
	FinTask(context.Context, *TaskRequest) (*TaskResponse, error)
}

// UnimplementedDispatcherSvrServer can be embedded to have forward compatible implementations.
type UnimplementedDispatcherSvrServer struct {
}

func (*UnimplementedDispatcherSvrServer) ReqTask(context.Context, *TaskRequest) (*TaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReqTask not implemented")
}
func (*UnimplementedDispatcherSvrServer) FinTask(context.Context, *TaskRequest) (*TaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinTask not implemented")
}

func RegisterDispatcherSvrServer(s *grpc.Server, srv DispatcherSvrServer) {
	s.RegisterService(&_DispatcherSvr_serviceDesc, srv)
}

func _DispatcherSvr_ReqTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DispatcherSvrServer).ReqTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.DispatcherSvr/ReqTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DispatcherSvrServer).ReqTask(ctx, req.(*TaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DispatcherSvr_FinTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DispatcherSvrServer).FinTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.DispatcherSvr/FinTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DispatcherSvrServer).FinTask(ctx, req.(*TaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DispatcherSvr_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.DispatcherSvr",
	HandlerType: (*DispatcherSvrServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "req_task",
			Handler:    _DispatcherSvr_ReqTask_Handler,
		},
		{
			MethodName: "fin_task",
			Handler:    _DispatcherSvr_FinTask_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "req.proto",
}
