// Code generated by protoc-gen-go-vtproto. DO NOT EDIT.
// protoc-gen-go-vtproto version: v0.6.0
// source: remote/remote.proto

package remote

import (
	context "context"
	fmt "fmt"
	actor "github.com/muhtutorials/actors/actor"
	protohelpers "github.com/planetscale/vtprotobuf/protohelpers"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	proto "google.golang.org/protobuf/proto"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	io "io"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

func (m *Envelope) CloneVT() *Envelope {
	if m == nil {
		return (*Envelope)(nil)
	}
	r := new(Envelope)
	if rhs := m.TypeNames; rhs != nil {
		tmpContainer := make([]string, len(rhs))
		copy(tmpContainer, rhs)
		r.TypeNames = tmpContainer
	}
	if rhs := m.Targets; rhs != nil {
		tmpContainer := make([]*actor.PID, len(rhs))
		for k, v := range rhs {
			if vtpb, ok := interface{}(v).(interface{ CloneVT() *actor.PID }); ok {
				tmpContainer[k] = vtpb.CloneVT()
			} else {
				tmpContainer[k] = proto.Clone(v).(*actor.PID)
			}
		}
		r.Targets = tmpContainer
	}
	if rhs := m.Senders; rhs != nil {
		tmpContainer := make([]*actor.PID, len(rhs))
		for k, v := range rhs {
			if vtpb, ok := interface{}(v).(interface{ CloneVT() *actor.PID }); ok {
				tmpContainer[k] = vtpb.CloneVT()
			} else {
				tmpContainer[k] = proto.Clone(v).(*actor.PID)
			}
		}
		r.Senders = tmpContainer
	}
	if rhs := m.Messages; rhs != nil {
		tmpContainer := make([]*Message, len(rhs))
		for k, v := range rhs {
			tmpContainer[k] = v.CloneVT()
		}
		r.Messages = tmpContainer
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Envelope) CloneMessageVT() proto.Message {
	return m.CloneVT()
}

func (m *Message) CloneVT() *Message {
	if m == nil {
		return (*Message)(nil)
	}
	r := new(Message)
	r.TypeNameIndex = m.TypeNameIndex
	r.TargetIndex = m.TargetIndex
	r.SenderIndex = m.SenderIndex
	if rhs := m.Data; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.Data = tmpBytes
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Message) CloneMessageVT() proto.Message {
	return m.CloneVT()
}

func (m *TestMessage) CloneVT() *TestMessage {
	if m == nil {
		return (*TestMessage)(nil)
	}
	r := new(TestMessage)
	if rhs := m.Data; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.Data = tmpBytes
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *TestMessage) CloneMessageVT() proto.Message {
	return m.CloneVT()
}

func (this *Envelope) EqualVT(that *Envelope) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if len(this.TypeNames) != len(that.TypeNames) {
		return false
	}
	for i, vx := range this.TypeNames {
		vy := that.TypeNames[i]
		if vx != vy {
			return false
		}
	}
	if len(this.Targets) != len(that.Targets) {
		return false
	}
	for i, vx := range this.Targets {
		vy := that.Targets[i]
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &actor.PID{}
			}
			if q == nil {
				q = &actor.PID{}
			}
			if equal, ok := interface{}(p).(interface{ EqualVT(*actor.PID) bool }); ok {
				if !equal.EqualVT(q) {
					return false
				}
			} else if !proto.Equal(p, q) {
				return false
			}
		}
	}
	if len(this.Senders) != len(that.Senders) {
		return false
	}
	for i, vx := range this.Senders {
		vy := that.Senders[i]
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &actor.PID{}
			}
			if q == nil {
				q = &actor.PID{}
			}
			if equal, ok := interface{}(p).(interface{ EqualVT(*actor.PID) bool }); ok {
				if !equal.EqualVT(q) {
					return false
				}
			} else if !proto.Equal(p, q) {
				return false
			}
		}
	}
	if len(this.Messages) != len(that.Messages) {
		return false
	}
	for i, vx := range this.Messages {
		vy := that.Messages[i]
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &Message{}
			}
			if q == nil {
				q = &Message{}
			}
			if !p.EqualVT(q) {
				return false
			}
		}
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Envelope) EqualMessageVT(thatMsg proto.Message) bool {
	that, ok := thatMsg.(*Envelope)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *Message) EqualVT(that *Message) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.TypeNameIndex != that.TypeNameIndex {
		return false
	}
	if this.TargetIndex != that.TargetIndex {
		return false
	}
	if this.SenderIndex != that.SenderIndex {
		return false
	}
	if string(this.Data) != string(that.Data) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Message) EqualMessageVT(thatMsg proto.Message) bool {
	that, ok := thatMsg.(*Message)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *TestMessage) EqualVT(that *TestMessage) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if string(this.Data) != string(that.Data) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *TestMessage) EqualMessageVT(thatMsg proto.Message) bool {
	that, ok := thatMsg.(*TestMessage)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RemoteClient is the client API for Remote service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RemoteClient interface {
	Receive(ctx context.Context, opts ...grpc.CallOption) (Remote_ReceiveClient, error)
}

type remoteClient struct {
	cc grpc.ClientConnInterface
}

func NewRemoteClient(cc grpc.ClientConnInterface) RemoteClient {
	return &remoteClient{cc}
}

func (c *remoteClient) Receive(ctx context.Context, opts ...grpc.CallOption) (Remote_ReceiveClient, error) {
	stream, err := c.cc.NewStream(ctx, &Remote_ServiceDesc.Streams[0], "/remote.Remote/Receive", opts...)
	if err != nil {
		return nil, err
	}
	x := &remoteReceiveClient{stream}
	return x, nil
}

type Remote_ReceiveClient interface {
	Send(*Envelope) error
	Recv() (*Envelope, error)
	grpc.ClientStream
}

type remoteReceiveClient struct {
	grpc.ClientStream
}

func (x *remoteReceiveClient) Send(m *Envelope) error {
	return x.ClientStream.SendMsg(m)
}

func (x *remoteReceiveClient) Recv() (*Envelope, error) {
	m := new(Envelope)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoteServer is the server API for Remote service.
// All implementations must embed UnimplementedRemoteServer
// for forward compatibility
type RemoteServer interface {
	Receive(Remote_ReceiveServer) error
	mustEmbedUnimplementedRemoteServer()
}

// UnimplementedRemoteServer must be embedded to have forward compatible implementations.
type UnimplementedRemoteServer struct {
}

func (UnimplementedRemoteServer) Receive(Remote_ReceiveServer) error {
	return status.Errorf(codes.Unimplemented, "method Receive not implemented")
}
func (UnimplementedRemoteServer) mustEmbedUnimplementedRemoteServer() {}

// UnsafeRemoteServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RemoteServer will
// result in compilation errors.
type UnsafeRemoteServer interface {
	mustEmbedUnimplementedRemoteServer()
}

func RegisterRemoteServer(s grpc.ServiceRegistrar, srv RemoteServer) {
	s.RegisterService(&Remote_ServiceDesc, srv)
}

func _Remote_Receive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RemoteServer).Receive(&remoteReceiveServer{stream})
}

type Remote_ReceiveServer interface {
	Send(*Envelope) error
	Recv() (*Envelope, error)
	grpc.ServerStream
}

type remoteReceiveServer struct {
	grpc.ServerStream
}

func (x *remoteReceiveServer) Send(m *Envelope) error {
	return x.ServerStream.SendMsg(m)
}

func (x *remoteReceiveServer) Recv() (*Envelope, error) {
	m := new(Envelope)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Remote_ServiceDesc is the grpc.ServiceDesc for Remote service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Remote_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "remote.Remote",
	HandlerType: (*RemoteServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Receive",
			Handler:       _Remote_Receive_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "remote/remote.proto",
}

func (m *Envelope) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Envelope) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Envelope) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Messages) > 0 {
		for iNdEx := len(m.Messages) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Messages[iNdEx].MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protohelpers.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Senders) > 0 {
		for iNdEx := len(m.Senders) - 1; iNdEx >= 0; iNdEx-- {
			if vtmsg, ok := interface{}(m.Senders[iNdEx]).(interface {
				MarshalToSizedBufferVT([]byte) (int, error)
			}); ok {
				size, err := vtmsg.MarshalToSizedBufferVT(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = protohelpers.EncodeVarint(dAtA, i, uint64(size))
			} else {
				encoded, err := proto.Marshal(m.Senders[iNdEx])
				if err != nil {
					return 0, err
				}
				i -= len(encoded)
				copy(dAtA[i:], encoded)
				i = protohelpers.EncodeVarint(dAtA, i, uint64(len(encoded)))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Targets) > 0 {
		for iNdEx := len(m.Targets) - 1; iNdEx >= 0; iNdEx-- {
			if vtmsg, ok := interface{}(m.Targets[iNdEx]).(interface {
				MarshalToSizedBufferVT([]byte) (int, error)
			}); ok {
				size, err := vtmsg.MarshalToSizedBufferVT(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = protohelpers.EncodeVarint(dAtA, i, uint64(size))
			} else {
				encoded, err := proto.Marshal(m.Targets[iNdEx])
				if err != nil {
					return 0, err
				}
				i -= len(encoded)
				copy(dAtA[i:], encoded)
				i = protohelpers.EncodeVarint(dAtA, i, uint64(len(encoded)))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.TypeNames) > 0 {
		for iNdEx := len(m.TypeNames) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.TypeNames[iNdEx])
			copy(dAtA[i:], m.TypeNames[iNdEx])
			i = protohelpers.EncodeVarint(dAtA, i, uint64(len(m.TypeNames[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Message) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Message) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Message) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = protohelpers.EncodeVarint(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0x22
	}
	if m.SenderIndex != 0 {
		i = protohelpers.EncodeVarint(dAtA, i, uint64(m.SenderIndex))
		i--
		dAtA[i] = 0x18
	}
	if m.TargetIndex != 0 {
		i = protohelpers.EncodeVarint(dAtA, i, uint64(m.TargetIndex))
		i--
		dAtA[i] = 0x10
	}
	if m.TypeNameIndex != 0 {
		i = protohelpers.EncodeVarint(dAtA, i, uint64(m.TypeNameIndex))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *TestMessage) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TestMessage) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *TestMessage) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = protohelpers.EncodeVarint(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Envelope) MarshalVTStrict() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVTStrict(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Envelope) MarshalToVTStrict(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVTStrict(dAtA[:size])
}

func (m *Envelope) MarshalToSizedBufferVTStrict(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Messages) > 0 {
		for iNdEx := len(m.Messages) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Messages[iNdEx].MarshalToSizedBufferVTStrict(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protohelpers.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Senders) > 0 {
		for iNdEx := len(m.Senders) - 1; iNdEx >= 0; iNdEx-- {
			if vtmsg, ok := interface{}(m.Senders[iNdEx]).(interface {
				MarshalToSizedBufferVTStrict([]byte) (int, error)
			}); ok {
				size, err := vtmsg.MarshalToSizedBufferVTStrict(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = protohelpers.EncodeVarint(dAtA, i, uint64(size))
			} else {
				encoded, err := proto.Marshal(m.Senders[iNdEx])
				if err != nil {
					return 0, err
				}
				i -= len(encoded)
				copy(dAtA[i:], encoded)
				i = protohelpers.EncodeVarint(dAtA, i, uint64(len(encoded)))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Targets) > 0 {
		for iNdEx := len(m.Targets) - 1; iNdEx >= 0; iNdEx-- {
			if vtmsg, ok := interface{}(m.Targets[iNdEx]).(interface {
				MarshalToSizedBufferVTStrict([]byte) (int, error)
			}); ok {
				size, err := vtmsg.MarshalToSizedBufferVTStrict(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = protohelpers.EncodeVarint(dAtA, i, uint64(size))
			} else {
				encoded, err := proto.Marshal(m.Targets[iNdEx])
				if err != nil {
					return 0, err
				}
				i -= len(encoded)
				copy(dAtA[i:], encoded)
				i = protohelpers.EncodeVarint(dAtA, i, uint64(len(encoded)))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.TypeNames) > 0 {
		for iNdEx := len(m.TypeNames) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.TypeNames[iNdEx])
			copy(dAtA[i:], m.TypeNames[iNdEx])
			i = protohelpers.EncodeVarint(dAtA, i, uint64(len(m.TypeNames[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Message) MarshalVTStrict() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVTStrict(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Message) MarshalToVTStrict(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVTStrict(dAtA[:size])
}

func (m *Message) MarshalToSizedBufferVTStrict(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = protohelpers.EncodeVarint(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0x22
	}
	if m.SenderIndex != 0 {
		i = protohelpers.EncodeVarint(dAtA, i, uint64(m.SenderIndex))
		i--
		dAtA[i] = 0x18
	}
	if m.TargetIndex != 0 {
		i = protohelpers.EncodeVarint(dAtA, i, uint64(m.TargetIndex))
		i--
		dAtA[i] = 0x10
	}
	if m.TypeNameIndex != 0 {
		i = protohelpers.EncodeVarint(dAtA, i, uint64(m.TypeNameIndex))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *TestMessage) MarshalVTStrict() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVTStrict(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TestMessage) MarshalToVTStrict(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVTStrict(dAtA[:size])
}

func (m *TestMessage) MarshalToSizedBufferVTStrict(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = protohelpers.EncodeVarint(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Envelope) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.TypeNames) > 0 {
		for _, s := range m.TypeNames {
			l = len(s)
			n += 1 + l + protohelpers.SizeOfVarint(uint64(l))
		}
	}
	if len(m.Targets) > 0 {
		for _, e := range m.Targets {
			if size, ok := interface{}(e).(interface {
				SizeVT() int
			}); ok {
				l = size.SizeVT()
			} else {
				l = proto.Size(e)
			}
			n += 1 + l + protohelpers.SizeOfVarint(uint64(l))
		}
	}
	if len(m.Senders) > 0 {
		for _, e := range m.Senders {
			if size, ok := interface{}(e).(interface {
				SizeVT() int
			}); ok {
				l = size.SizeVT()
			} else {
				l = proto.Size(e)
			}
			n += 1 + l + protohelpers.SizeOfVarint(uint64(l))
		}
	}
	if len(m.Messages) > 0 {
		for _, e := range m.Messages {
			l = e.SizeVT()
			n += 1 + l + protohelpers.SizeOfVarint(uint64(l))
		}
	}
	n += len(m.unknownFields)
	return n
}

func (m *Message) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.TypeNameIndex != 0 {
		n += 1 + protohelpers.SizeOfVarint(uint64(m.TypeNameIndex))
	}
	if m.TargetIndex != 0 {
		n += 1 + protohelpers.SizeOfVarint(uint64(m.TargetIndex))
	}
	if m.SenderIndex != 0 {
		n += 1 + protohelpers.SizeOfVarint(uint64(m.SenderIndex))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + protohelpers.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *TestMessage) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + protohelpers.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *Envelope) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protohelpers.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Envelope: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Envelope: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeNames", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TypeNames = append(m.TypeNames, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Targets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Targets = append(m.Targets, &actor.PID{})
			if unmarshal, ok := interface{}(m.Targets[len(m.Targets)-1]).(interface {
				UnmarshalVT([]byte) error
			}); ok {
				if err := unmarshal.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
					return err
				}
			} else {
				if err := proto.Unmarshal(dAtA[iNdEx:postIndex], m.Targets[len(m.Targets)-1]); err != nil {
					return err
				}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Senders", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Senders = append(m.Senders, &actor.PID{})
			if unmarshal, ok := interface{}(m.Senders[len(m.Senders)-1]).(interface {
				UnmarshalVT([]byte) error
			}); ok {
				if err := unmarshal.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
					return err
				}
			} else {
				if err := proto.Unmarshal(dAtA[iNdEx:postIndex], m.Senders[len(m.Senders)-1]); err != nil {
					return err
				}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Messages", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Messages = append(m.Messages, &Message{})
			if err := m.Messages[len(m.Messages)-1].UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protohelpers.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protohelpers.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Message) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protohelpers.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Message: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Message: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeNameIndex", wireType)
			}
			m.TypeNameIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TypeNameIndex |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetIndex", wireType)
			}
			m.TargetIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TargetIndex |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SenderIndex", wireType)
			}
			m.SenderIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SenderIndex |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protohelpers.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protohelpers.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TestMessage) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protohelpers.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TestMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TestMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protohelpers.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protohelpers.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Envelope) UnmarshalVTUnsafe(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protohelpers.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Envelope: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Envelope: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeNames", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var stringValue string
			if intStringLen > 0 {
				stringValue = unsafe.String(&dAtA[iNdEx], intStringLen)
			}
			m.TypeNames = append(m.TypeNames, stringValue)
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Targets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Targets = append(m.Targets, &actor.PID{})
			if unmarshal, ok := interface{}(m.Targets[len(m.Targets)-1]).(interface {
				UnmarshalVTUnsafe([]byte) error
			}); ok {
				if err := unmarshal.UnmarshalVTUnsafe(dAtA[iNdEx:postIndex]); err != nil {
					return err
				}
			} else {
				if err := proto.Unmarshal(dAtA[iNdEx:postIndex], m.Targets[len(m.Targets)-1]); err != nil {
					return err
				}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Senders", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Senders = append(m.Senders, &actor.PID{})
			if unmarshal, ok := interface{}(m.Senders[len(m.Senders)-1]).(interface {
				UnmarshalVTUnsafe([]byte) error
			}); ok {
				if err := unmarshal.UnmarshalVTUnsafe(dAtA[iNdEx:postIndex]); err != nil {
					return err
				}
			} else {
				if err := proto.Unmarshal(dAtA[iNdEx:postIndex], m.Senders[len(m.Senders)-1]); err != nil {
					return err
				}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Messages", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Messages = append(m.Messages, &Message{})
			if err := m.Messages[len(m.Messages)-1].UnmarshalVTUnsafe(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protohelpers.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protohelpers.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Message) UnmarshalVTUnsafe(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protohelpers.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Message: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Message: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeNameIndex", wireType)
			}
			m.TypeNameIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TypeNameIndex |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetIndex", wireType)
			}
			m.TargetIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TargetIndex |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SenderIndex", wireType)
			}
			m.SenderIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SenderIndex |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = dAtA[iNdEx:postIndex]
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protohelpers.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protohelpers.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TestMessage) UnmarshalVTUnsafe(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protohelpers.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TestMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TestMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protohelpers.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return protohelpers.ErrInvalidLength
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return protohelpers.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = dAtA[iNdEx:postIndex]
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protohelpers.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protohelpers.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}