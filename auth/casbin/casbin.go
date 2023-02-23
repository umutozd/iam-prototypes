package casbin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/casbin/casbin"
	"github.com/sirupsen/logrus"
	"github.com/umutozd/iam-prototypes/auth"
)

type casbinAuth struct {
	enforcer *casbin.Enforcer
	adapter  *readOnlyAdapter
}

// NewCasbinAuth initializes a new casbin AuthLib with the given permission and inheritance rules.
// Permission rules must be in the form:
//
//	[]string{"role", "resource", "permission"}
//
// The inheritance rules must be in the form:
//
//	[]string{"inheritor", "inherited"}
func NewCasbinAuth(permissionRules, inheritanceRules [][]string) (auth.AuthLib, error) {
	// check the structure of the rules
	for i, rule := range permissionRules {
		if l := len(rule); l != casbinPermissionRuleSize {
			return nil, fmt.Errorf("casbin: permission at index %d has %d elements, but must have %d", i, l, casbinPermissionRuleSize)
		}
	}
	for i, rule := range inheritanceRules {
		if l := len(rule); l != casbinInheritanceRuleSize {
			return nil, fmt.Errorf("casbin: inheritance at index %d has %d elements, but must have %d", i, l, casbinInheritanceRuleSize)
		}
	}

	// create the internal enforces using model string and custom adapter
	enf := &casbin.Enforcer{}
	model := casbin.NewModel(casbinModel)
	adapter := &readOnlyAdapter{
		permissions:      permissionRules,
		inheritanceRules: inheritanceRules,
	}
	enf.InitWithModelAndAdapter(model, adapter)

	return &casbinAuth{
		enforcer: enf,
		adapter:  adapter,
	}, nil
}

func (ca *casbinAuth) Name() auth.AuthLibName {
	return auth.AuthLibNameCasbin
}

func (ca *casbinAuth) Authorize(userId string, usergroups []string, resource any, action string) bool {
	ca.adapter.mu.Lock()
	defer ca.adapter.mu.Unlock()

	groups := append([]string{}, usergroups...)
	clonedUserId := strings.Clone(userId)
	ca.adapter.loaderFunc = func() (userId string, usergroups []string) {
		return clonedUserId, groups
	}
	if err := ca.enforcer.LoadPolicy(); err != nil {
		logrus.WithError(err).Error("casbin: error reloading policy for Authorize call")
		return false
	}

	logrus.Infof("casbin: enforcing user=%s, resource=%s, action=%s", userId, ca.getResourceName(resource), action)
	allowed, err := ca.enforcer.EnforceSafe(userId, ca.getResourceName(resource), action)
	if err != nil {
		logrus.WithError(err).Error("casbin: enforce returned error")
	}
	return allowed
}

func (ca *casbinAuth) getResourceName(resource any) string {
	if str, ok := resource.(string); ok {
		return str
	}
	if t := reflect.TypeOf(resource); t.Kind() == reflect.Pointer {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
