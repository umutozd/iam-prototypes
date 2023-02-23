package casbin

// casbinModel is the model string for the Casbin auth library. This part can be a constant
// in the plugin code because we'll most probably won't change this.
const casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

const (
	casbinPermissionRuleSize  = 3
	casbinInheritanceRuleSize = 2
)
