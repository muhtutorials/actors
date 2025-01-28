package remote

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

var registry = make(map[string]VTUnmarshaler)

func RegisterType(v VTUnmarshaler) {
	name := string(proto.MessageName(v))
	registry[name] = v
}

func GetType(t string) (VTUnmarshaler, error) {
	if m, ok := registry[t]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("type (%s) is not registered. Make sure to register it with 'remote.RegisterType(&instance{})'", t)
}
