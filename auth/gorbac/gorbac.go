package gorbac

import (
	"fmt"

	"github.com/mikespook/gorbac"
	"github.com/umutozd/iam-prototypes/auth"
)

type gorbacAuth struct {
	gorbac      *gorbac.RBAC
	roles       map[string]*gorbac.StdRole
	permissions map[string]gorbac.Permission
}

func NewGorbacAuth(roles []string, permissions []string, permToRoles map[string][]string, inheritanceRules map[string][]string) (auth.AuthLib, error) {
	roleMap := make(map[string]*gorbac.StdRole, len(roles))
	permMap := make(map[string]gorbac.Permission, len(permissions))

	// define roles, permissions and assign permissions to roles
	for _, roleName := range roles {
		role := gorbac.NewStdRole(roleName)
		roleMap[roleName] = role
	}
	for _, permName := range permissions {
		perm := gorbac.NewStdPermission(permName)
		permMap[permName] = perm
	}
	for permName, roleNames := range permToRoles {
		perm, ok := permMap[permName]
		if !ok {
			return nil, fmt.Errorf("gorbac: invalid mapping of permission to role: permission %s doesn't exist", permName)
		}
		for _, roleName := range roleNames {
			role, ok := roleMap[roleName]
			if !ok {
				return nil, fmt.Errorf("gorbac: invalid mapping of permission to role: role %s doesn't exist", roleName)
			}
			if err := role.Assign(perm); err != nil {
				return nil, fmt.Errorf("gorbac: error assigning permission %s to role %s: %v", permName, roleName, err)
			}
		}
	}

	// add the created roles to a new gorbac instance
	rbac := gorbac.New()
	for roleName, role := range roleMap {
		if err := rbac.Add(role); err != nil {
			return nil, fmt.Errorf("gorbac: error adding role %s to internal gorbac instance: %v", roleName, err)
		}
	}

	// create inheritance relationships between roles
	for inheritorName, inheritedNames := range inheritanceRules {
		_, ok := roleMap[inheritorName]
		if !ok {
			return nil, fmt.Errorf("gorbac: inheritor role %s doesn't exist", inheritorName)
		}
		inheritedIds := make([]string, 0, len(inheritedNames))
		for _, inheritedName := range inheritedNames {
			if inheritedName == inheritorName {
				// inheriting from self is no-op and it causes gorbac to recurse infinitely
				continue
			}
			_, ok := roleMap[inheritedName]
			if !ok {
				return nil, fmt.Errorf("gorbac: inherited role %s doesn't exist", inheritedName)
			}

			inheritedIds = append(inheritedIds, inheritedName)
		}

		if err := rbac.SetParents(inheritorName, inheritedIds); err != nil {
			return nil, fmt.Errorf("gorbac: error setting inheritance relations for inheritor %s: %v", inheritorName, err)
		}
	}

	return &gorbacAuth{
		gorbac:      rbac,
		roles:       roleMap,
		permissions: permMap,
	}, nil
}

func (g *gorbacAuth) Name() auth.AuthLibName {
	return auth.AuthLibNameGorbac
}

func (g *gorbacAuth) Authorize(_ string, usergroups []string, _ any, action string) bool {
	for _, group := range usergroups {
		role, ok := g.roles[group]
		if !ok {
			continue
		}
		perm, ok := g.permissions[action]
		if !ok {
			continue
		}
		if granted := g.gorbac.IsGranted(role.ID(), perm, nil); granted {
			return true
		}
	}
	return false
}
