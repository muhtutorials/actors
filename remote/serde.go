package remote

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type Serializer interface {
	Serialize(any) ([]byte, error)
	TypeName(any) string
}

type Deserializer interface {
	Deserialize([]byte, string) (any, error)
}

type VTMarshaler interface {
	proto.Message
	MarshalVT() ([]byte, error)
}

type VTUnmarshaler interface {
	proto.Message
	UnmarshalVT([]byte) error
}

type ProtoSerde struct{}

func (ProtoSerde) Serialize(msg any) ([]byte, error) {
	return proto.Marshal(msg.(proto.Message))
}

func (ProtoSerde) TypeName(msg any) string {
	return string(proto.MessageName(msg.(proto.Message)))
}

func (ProtoSerde) Deserialize(data []byte, typeName string) (any, error) {
	name := protoreflect.FullName(typeName)
	messageType, err := protoregistry.GlobalTypes.FindMessageByName(name)
	if err != nil {
		return nil, err
	}
	protoMessage := messageType.New().Interface()
	err = proto.Unmarshal(data, protoMessage)
	return protoMessage, err
}

// todo: delete if not used

type VTProtoSerde struct{}

func (VTProtoSerde) Serialize(msg any) ([]byte, error) {
	return msg.(VTMarshaler).MarshalVT()
}

func (VTProtoSerde) TypeName(msg any) string {
	return string(proto.MessageName(msg.(proto.Message)))
}

func (VTProtoSerde) Deserialize(data []byte, typeName string) (any, error) {
	v, err := GetType(typeName)
	if err != nil {
		return nil, err
	}
	return v, v.UnmarshalVT(data)
}
