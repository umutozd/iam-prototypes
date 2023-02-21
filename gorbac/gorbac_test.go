package gorbac

import (
	"testing"

	"github.com/mikespook/gorbac"
)

func TestGorbac(t *testing.T) {
	// role definitions
	ticketViewer := gorbac.NewStdRole("otsimo.com/ticket/viewer")
	ticketEditor := gorbac.NewStdRole("otsimo.com/ticket/editor")
	ticketMaintainer := gorbac.NewStdRole("otsimo.com/ticket/maintainer")
	someUnaddedRole := gorbac.NewStdRole("otsimo.com/invalid/role")

	// permission definitions
	permRead := gorbac.NewStdPermission("read")
	permWrite := gorbac.NewStdPermission("write")
	permDelete := gorbac.NewStdPermission("delete")

	// assign permissions to roles
	mustAssignRole(t, ticketViewer, permRead)
	mustAssignRole(t, ticketEditor, permWrite)
	mustAssignRole(t, ticketMaintainer, permDelete)

	// add roles to the RBAC instance
	rbac := gorbac.New()
	mustAddRole(t, rbac, ticketViewer)
	mustAddRole(t, rbac, ticketEditor)
	mustAddRole(t, rbac, ticketMaintainer)

	// set inheritance rules: child has all permissions of its parents
	mustSetParents(t, rbac, ticketEditor, ticketViewer)
	mustSetParents(t, rbac, ticketMaintainer, ticketEditor, ticketViewer)

	cases := []*struct {
		name     string
		role     gorbac.Role
		perm     gorbac.Permission
		expected bool
	}{
		{
			name:     "viewer can read",
			role:     ticketViewer,
			perm:     permRead,
			expected: true,
		},
		{
			name:     "viewer can't write",
			role:     ticketViewer,
			perm:     permWrite,
			expected: false,
		},
		{
			name:     "viewer can't delete",
			role:     ticketViewer,
			perm:     permDelete,
			expected: false,
		},
		{
			name:     "editor can read",
			role:     ticketEditor,
			perm:     permRead,
			expected: true,
		},
		{
			name:     "editor can write",
			role:     ticketEditor,
			perm:     permWrite,
			expected: true,
		},
		{
			name:     "editor can't delete",
			role:     ticketEditor,
			perm:     permDelete,
			expected: false,
		},
		{
			name:     "maintainer can read",
			role:     ticketMaintainer,
			perm:     permRead,
			expected: true,
		},
		{
			name:     "maintainer can write",
			role:     ticketMaintainer,
			perm:     permWrite,
			expected: true,
		},
		{
			name:     "maintainer can delete",
			role:     ticketMaintainer,
			perm:     permDelete,
			expected: true,
		},
		{
			name:     "invalid role can't read",
			role:     someUnaddedRole,
			perm:     permRead,
			expected: false,
		},
		{
			name:     "invalid role can't write",
			role:     someUnaddedRole,
			perm:     permWrite,
			expected: false,
		},
		{
			name:     "invalid role can't delete",
			role:     someUnaddedRole,
			perm:     permDelete,
			expected: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			granted := rbac.IsGranted(c.role.ID(), c.perm, nil)
			if granted != c.expected {
				t.FailNow()
			}
		})
	}
}

func mustAssignRole(t *testing.T, role *gorbac.StdRole, perm gorbac.Permission) {
	if err := role.Assign(perm); err != nil {
		t.Fatalf("error assigning permission %q to role %q: %v", perm.ID(), role.ID(), err)
	}
}

func mustAddRole(t *testing.T, rbac *gorbac.RBAC, role gorbac.Role) {
	if err := rbac.Add(role); err != nil {
		t.Fatalf("error adding role %q to rbac instance: %v", role.ID(), err)
	}
}

func mustSetParents(t *testing.T, rbac *gorbac.RBAC, childRole gorbac.Role, parentRoles ...gorbac.Role) {
	parentIds := make([]string, 0, len(parentRoles))
	for _, pr := range parentRoles {
		parentIds = append(parentIds, pr.ID())
	}
	if err := rbac.SetParents(childRole.ID(), parentIds); err != nil {
		t.Fatalf("error setting parents %+v to role %q: %v", parentIds, childRole.ID(), err)
	}
}
