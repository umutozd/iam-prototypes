package otsimo

import (
	proto "github.com/golang/protobuf/proto" //nolint
	google_protobuf "google.golang.org/protobuf/types/descriptorpb"
)

func GetAuthConfig(method *google_protobuf.MethodDescriptorProto) *OtsimoAuth {
	if method == nil {
		return nil
	}
	if method.Options != nil {
		v, err := proto.GetExtension(method.Options, E_Config)
		if err == nil && v.(*OtsimoAuth) != nil {
			return (v.(*OtsimoAuth))
		}
	}
	return nil
}
