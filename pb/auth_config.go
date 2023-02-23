package pb

import "github.com/golang/protobuf/proto"

// OsoPolicy is the policy string for Oso auth library. This part should be generated
// automatically from rpc options in .proto files.
const OsoPolicy = `
allow(actor, action, resource) if
has_permission(actor, action, resource);

actor User {}

resource Foo {
	permissions = ["read", "write", "delete"];
	roles = [
		"otsimo.com/foo/viewer",
		"otsimo.com/foo/editor",
		"otsimo.com/foo/maintainer"
	];

	"read" if "otsimo.com/foo/viewer";
	"write" if "otsimo.com/foo/editor";
	"delete" if "otsimo.com/foo/maintainer";

	"otsimo.com/foo/editor" if "otsimo.com/foo/maintainer";
	"otsimo.com/foo/viewer" if "otsimo.com/foo/editor";
}

has_role(user: User, roleName: String, _: Foo) if
role in user.Roles and
role = roleName;
`

type ServiceAuthConfig struct {
	ServiceName string
	Methods     []*AuthMethodConfig
}

type AuthMethodConfig struct {
	FullMethod  string
	Resource    proto.Message
	Action      string
	AllowAPIKey bool
}

// SimpleServiceAuthConfig is the auth config to be used by the new auth libraries. This should
// always be generated bu a protoc plugin.
var SimpleServiceAuthConfig = ServiceAuthConfig{
	ServiceName: "SimpleService",
	Methods: []*AuthMethodConfig{
		{
			FullMethod: "/otsimo.simple.v1.SimpleService/GetFoo",
			Resource:   &Foo{},
			Action:     "read",
		},
		{
			FullMethod: "/otsimo.simple.v1.SimpleService/UpdateFoo",
			Resource:   &Foo{},
			Action:     "write",
		},
	},
}
