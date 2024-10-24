// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.1
// source: proto/minor-upgrade.proto

package proto

import (
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

type PrerequisiteCheck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AvailableSpace bool `protobuf:"varint,1,opt,name=availableSpace,proto3" json:"availableSpace,omitempty"`
	// todo: try from the lighthouse agent
	EtcdStatus       bool   `protobuf:"varint,2,opt,name=etcdStatus,proto3" json:"etcdStatus,omitempty"`
	AvailableStorage uint64 `protobuf:"varint,3,opt,name=availableStorage,proto3" json:"availableStorage,omitempty"`
}

func (x *PrerequisiteCheck) Reset() {
	*x = PrerequisiteCheck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_minor_upgrade_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PrerequisiteCheck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PrerequisiteCheck) ProtoMessage() {}

func (x *PrerequisiteCheck) ProtoReflect() protoreflect.Message {
	mi := &file_proto_minor_upgrade_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PrerequisiteCheck.ProtoReflect.Descriptor instead.
func (*PrerequisiteCheck) Descriptor() ([]byte, []int) {
	return file_proto_minor_upgrade_proto_rawDescGZIP(), []int{0}
}

func (x *PrerequisiteCheck) GetAvailableSpace() bool {
	if x != nil {
		return x.AvailableSpace
	}
	return false
}

func (x *PrerequisiteCheck) GetEtcdStatus() bool {
	if x != nil {
		return x.EtcdStatus
	}
	return false
}

func (x *PrerequisiteCheck) GetAvailableStorage() uint64 {
	if x != nil {
		return x.AvailableStorage
	}
	return 0
}

type CreateUpgradeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrerequisiteCheckSuccess bool   `protobuf:"varint,1,opt,name=prerequisiteCheckSuccess,proto3" json:"prerequisiteCheckSuccess,omitempty"`
	UpgradeSuccess           bool   `protobuf:"varint,2,opt,name=upgradeSuccess,proto3" json:"upgradeSuccess,omitempty"`
	RestartSuccess           bool   `protobuf:"varint,3,opt,name=restartSuccess,proto3" json:"restartSuccess,omitempty"`
	RetryAttempt             uint32 `protobuf:"varint,4,opt,name=retryAttempt,proto3" json:"retryAttempt,omitempty"`
	Err                      string `protobuf:"bytes,5,opt,name=err,proto3" json:"err,omitempty"`
	Log                      string `protobuf:"bytes,6,opt,name=log,proto3" json:"log,omitempty"`
}

func (x *CreateUpgradeRequest) Reset() {
	*x = CreateUpgradeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_minor_upgrade_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateUpgradeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateUpgradeRequest) ProtoMessage() {}

func (x *CreateUpgradeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_minor_upgrade_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateUpgradeRequest.ProtoReflect.Descriptor instead.
func (*CreateUpgradeRequest) Descriptor() ([]byte, []int) {
	return file_proto_minor_upgrade_proto_rawDescGZIP(), []int{1}
}

func (x *CreateUpgradeRequest) GetPrerequisiteCheckSuccess() bool {
	if x != nil {
		return x.PrerequisiteCheckSuccess
	}
	return false
}

func (x *CreateUpgradeRequest) GetUpgradeSuccess() bool {
	if x != nil {
		return x.UpgradeSuccess
	}
	return false
}

func (x *CreateUpgradeRequest) GetRestartSuccess() bool {
	if x != nil {
		return x.RestartSuccess
	}
	return false
}

func (x *CreateUpgradeRequest) GetRetryAttempt() uint32 {
	if x != nil {
		return x.RetryAttempt
	}
	return 0
}

func (x *CreateUpgradeRequest) GetErr() string {
	if x != nil {
		return x.Err
	}
	return ""
}

func (x *CreateUpgradeRequest) GetLog() string {
	if x != nil {
		return x.Log
	}
	return ""
}

type CreateUpgradeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProceedNextStep      bool `protobuf:"varint,1,opt,name=proceedNextStep,proto3" json:"proceedNextStep,omitempty"`
	SkipRetryCurrentStep bool `protobuf:"varint,2,opt,name=skipRetryCurrentStep,proto3" json:"skipRetryCurrentStep,omitempty"`
	TerminateApplication bool `protobuf:"varint,3,opt,name=terminateApplication,proto3" json:"terminateApplication,omitempty"`
}

func (x *CreateUpgradeResponse) Reset() {
	*x = CreateUpgradeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_minor_upgrade_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateUpgradeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateUpgradeResponse) ProtoMessage() {}

func (x *CreateUpgradeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_minor_upgrade_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateUpgradeResponse.ProtoReflect.Descriptor instead.
func (*CreateUpgradeResponse) Descriptor() ([]byte, []int) {
	return file_proto_minor_upgrade_proto_rawDescGZIP(), []int{2}
}

func (x *CreateUpgradeResponse) GetProceedNextStep() bool {
	if x != nil {
		return x.ProceedNextStep
	}
	return false
}

func (x *CreateUpgradeResponse) GetSkipRetryCurrentStep() bool {
	if x != nil {
		return x.SkipRetryCurrentStep
	}
	return false
}

func (x *CreateUpgradeResponse) GetTerminateApplication() bool {
	if x != nil {
		return x.TerminateApplication
	}
	return false
}

var File_proto_minor_upgrade_proto protoreflect.FileDescriptor

var file_proto_minor_upgrade_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x69, 0x6e, 0x6f, 0x72, 0x2d, 0x75, 0x70,
	0x67, 0x72, 0x61, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x87, 0x01, 0x0a, 0x11,
	0x50, 0x72, 0x65, 0x72, 0x65, 0x71, 0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x70,
	0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x61, 0x76, 0x61, 0x69, 0x6c,
	0x61, 0x62, 0x6c, 0x65, 0x53, 0x70, 0x61, 0x63, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x74, 0x63,
	0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x65,
	0x74, 0x63, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2a, 0x0a, 0x10, 0x61, 0x76, 0x61,
	0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x10, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x22, 0xea, 0x01, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x55, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3a,
	0x0a, 0x18, 0x70, 0x72, 0x65, 0x72, 0x65, 0x71, 0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x18, 0x70, 0x72, 0x65, 0x72, 0x65, 0x71, 0x75, 0x69, 0x73, 0x69, 0x74, 0x65, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x75, 0x70,
	0x67, 0x72, 0x61, 0x64, 0x65, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0e, 0x75, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65, 0x53, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x72, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74, 0x53, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x72, 0x65, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65,
	0x74, 0x72, 0x79, 0x41, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x65, 0x72, 0x72,
	0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6c,
	0x6f, 0x67, 0x22, 0xa9, 0x01, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x55, 0x70, 0x67,
	0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x0f,
	0x70, 0x72, 0x6f, 0x63, 0x65, 0x65, 0x64, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x74, 0x65, 0x70, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x65, 0x64, 0x4e, 0x65,
	0x78, 0x74, 0x53, 0x74, 0x65, 0x70, 0x12, 0x32, 0x0a, 0x14, 0x73, 0x6b, 0x69, 0x70, 0x52, 0x65,
	0x74, 0x72, 0x79, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x65, 0x70, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x14, 0x73, 0x6b, 0x69, 0x70, 0x52, 0x65, 0x74, 0x72, 0x79, 0x43,
	0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x65, 0x70, 0x12, 0x32, 0x0a, 0x14, 0x74, 0x65,
	0x72, 0x6d, 0x69, 0x6e, 0x61, 0x74, 0x65, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x14, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e,
	0x61, 0x74, 0x65, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x32, 0x48,
	0x0a, 0x07, 0x55, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x15, 0x2e, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x55, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x16, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x55, 0x70, 0x67, 0x72, 0x61, 0x64, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x68, 0x69, 0x73, 0x68, 0x69, 0x72, 0x39, 0x31,
	0x35, 0x39, 0x2f, 0x6b, 0x61, 0x70, 0x65, 0x74, 0x61, 0x6e, 0x69, 0x6f, 0x73, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_minor_upgrade_proto_rawDescOnce sync.Once
	file_proto_minor_upgrade_proto_rawDescData = file_proto_minor_upgrade_proto_rawDesc
)

func file_proto_minor_upgrade_proto_rawDescGZIP() []byte {
	file_proto_minor_upgrade_proto_rawDescOnce.Do(func() {
		file_proto_minor_upgrade_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_minor_upgrade_proto_rawDescData)
	})
	return file_proto_minor_upgrade_proto_rawDescData
}

var file_proto_minor_upgrade_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_proto_minor_upgrade_proto_goTypes = []any{
	(*PrerequisiteCheck)(nil),     // 0: PrerequisiteCheck
	(*CreateUpgradeRequest)(nil),  // 1: CreateUpgradeRequest
	(*CreateUpgradeResponse)(nil), // 2: CreateUpgradeResponse
}
var file_proto_minor_upgrade_proto_depIdxs = []int32{
	1, // 0: Upgrade.StatusUpdate:input_type -> CreateUpgradeRequest
	2, // 1: Upgrade.StatusUpdate:output_type -> CreateUpgradeResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_minor_upgrade_proto_init() }
func file_proto_minor_upgrade_proto_init() {
	if File_proto_minor_upgrade_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_minor_upgrade_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*PrerequisiteCheck); i {
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
		file_proto_minor_upgrade_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*CreateUpgradeRequest); i {
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
		file_proto_minor_upgrade_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*CreateUpgradeResponse); i {
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
			RawDescriptor: file_proto_minor_upgrade_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_minor_upgrade_proto_goTypes,
		DependencyIndexes: file_proto_minor_upgrade_proto_depIdxs,
		MessageInfos:      file_proto_minor_upgrade_proto_msgTypes,
	}.Build()
	File_proto_minor_upgrade_proto = out.File
	file_proto_minor_upgrade_proto_rawDesc = nil
	file_proto_minor_upgrade_proto_goTypes = nil
	file_proto_minor_upgrade_proto_depIdxs = nil
}
