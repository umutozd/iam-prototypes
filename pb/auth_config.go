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

	CasbinConfig *CasbinConfig
	GorbacConfig *GorbacConfig
}

type AuthMethodConfig struct {
	FullMethod  string
	AllowAPIKey bool
	Resource    proto.Message
	Action      string
}

type CasbinConfig struct {
	Permissions      [][]string
	InheritanceRules [][]string
}

type GorbacConfig struct {
	Roles              []string
	Permissions        []string
	PermissionsToRoles map[string][]string
	InheritanceRules   map[string][]string
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
	CasbinConfig: &CasbinConfig{
		Permissions: [][]string{
			{"otsimo.com/foo/viewer", "Foo", "read"},
			{"otsimo.com/foo/editor", "Foo", "write"},
			{"otsimo.com/foo/maintainer", "Foo", "delete"},
		},
		InheritanceRules: [][]string{
			{"otsimo.com/foo/editor", "otsimo.com/foo/viewer"},
			{"otsimo.com/foo/maintainer", "otsimo.com/foo/editor"},
		},
	},
	GorbacConfig: &GorbacConfig{
		Roles: []string{
			"otsimo.com/foo/viewer",
			"otsimo.com/foo/editor",
			"otsimo.com/foo/maintainer",
		},
		Permissions: []string{"read", "write", "delete"},
		PermissionsToRoles: map[string][]string{
			"read":   {"otsimo.com/foo/viewer"},
			"write":  {"otsimo.com/foo/editor"},
			"delete": {"otsimo.com/foo/maintainer"},
		},
		InheritanceRules: map[string][]string{
			"otsimo.com/foo/maintainer": {
				"otsimo.com/foo/editor",
				"otsimo.com/foo/viewer",
			},
			"otsimo.com/foo/editor": {
				"otsimo.com/foo/viewer",
			},
		},
	},
}
