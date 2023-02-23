package casbin

import (
	"sync"

	"github.com/casbin/casbin/model"
	"github.com/sirupsen/logrus"
)

type readOnlyAdapter struct {
	// permissions are basic role-to-permission-on-resource items to be loaded into the policy.
	// The exact order must be: []string{"role", "resource", "permission"}
	permissions [][]string

	// inheritanceRules are role-to-role items that indicate a role inheriting the permissions of the other.
	// The exact order must be: []string{"inheritor", "inherited"}
	inheritanceRules [][]string

	// loaderFunc is a function that is used by LoadPolicy to load a user's group information to casbin
	loaderFunc func() (userId string, usergroups []string)

	// mutex protects the entire object
	mu sync.Mutex
}

func (a *readOnlyAdapter) LoadPolicy(m model.Model) error {
	// load permissions and inheritance rules
	for _, rule := range a.permissions {
		_ = m.AddPolicy("p", "p", rule)
	}
	for _, rule := range a.inheritanceRules {
		_ = m.AddPolicy("g", "g", rule)
	}

	// load policies of the current user, if loader func exists
	if a.loaderFunc == nil {
		logrus.Info("LoadPolicy: loaderFunc is nil, skipping")
		return nil
	}
	userId, usergroups := a.loaderFunc()
	logrus.Debugf("adapter: user has %d groups", len(usergroups))
	for _, ug := range usergroups {
		_ = m.AddPolicy("g", "g", []string{userId, ug})
	}
	return nil
}

// SavePolicy is no-op because this adapter is read-only.
func (a *readOnlyAdapter) SavePolicy(model model.Model) error {
	return nil
}

// AddPolicy is no-op because this adapter is read-only.
func (a *readOnlyAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemovePolicy is no-op because this adapter is read-only.
func (a *readOnlyAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy is no-op because this adapter is read-only.
func (a *readOnlyAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}
