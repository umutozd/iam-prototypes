package oso

import (
	"errors"
	"testing"

	"github.com/osohq/go-oso"
)

type Ticket struct {
	Id   int
	Body string
	Type string
}

type User struct {
	Id    int
	Name  string
	Roles []string
}

const (
	roleTicketViewer     = "otsimo.com/ticket/viewer"
	roleTicketEditor     = "otsimo.com/ticket/editor"
	roleTicketMaintainer = "otsimo.com/ticket/maintainer"
)

var ErrNotFound = errors.New("oso: user or access rule not found")
var ErrForbidden = errors.New("oso: user is forbidden to perform this action")

func TestOsoIAM(t *testing.T) {
	o, err := oso.NewOso()
	if err != nil {
		t.Fatalf("error creating an oso instance: %v", err)
	}
	// By setting the name of the "read" action, we can force oso to return a not-found error
	// when an actor doesn't have the permissions to even read the data. This is to protect any data.
	o.SetReadAction("read")
	o.SetNotFoundError(func() error { return ErrNotFound })
	o.SetForbiddenError(func() error { return ErrForbidden })

	if err = o.RegisterClass(Ticket{}, nil); err != nil {
		t.Fatalf("error registering Ticket type to oso: %v", err)
	}
	if err = o.RegisterClass(User{}, nil); err != nil {
		t.Fatalf("error registering User type to oso: %v", err)
	}

	if err = o.LoadFiles([]string{"main.polar"}); err != nil {
		t.Fatalf("error loading .polar file: %v", err)
	}

	cases := []*struct {
		name          string
		user          User
		ticket        Ticket
		action        string
		expectedError error
	}{
		{
			name: "viewer-read-allowed",
			user: User{
				Id:   1,
				Name: "user 1",
				Roles: []string{
					roleTicketViewer,
				},
			},
			ticket: Ticket{
				Id:   1,
				Body: "lorem ipsum",
				Type: "general",
			},
			action:        "read",
			expectedError: nil,
		},
		{
			name: "invalid role-read-not found",
			user: User{
				Id:   2,
				Name: "user 2",
				Roles: []string{
					"some other role",
				},
			},
			ticket: Ticket{
				Id:   2,
				Body: "lorem ipsum",
				Type: "general",
			},
			action:        "read",
			expectedError: ErrNotFound,
		},
		{
			name: "invalid role-write-forbidden",
			user: User{
				Id:   2,
				Name: "user 2",
				Roles: []string{
					"some other role",
				},
			},
			ticket: Ticket{
				Id:   2,
				Body: "lorem ipsum",
				Type: "general",
			},
			action:        "write",
			expectedError: ErrForbidden,
		},
		{
			name: "editor-write-allowed",
			user: User{
				Id:   3,
				Name: "user 3",
				Roles: []string{
					roleTicketEditor,
				},
			},
			ticket: Ticket{
				Id:   2,
				Body: "lorem ipsum",
				Type: "general",
			},
			action:        "write",
			expectedError: nil,
		},
		{
			name: "editor-delete-forbidden ",
			user: User{
				Id:   3,
				Name: "user 3",
				Roles: []string{
					roleTicketEditor,
				},
			},
			ticket: Ticket{
				Id:   2,
				Body: "lorem ipsum",
				Type: "general",
			},
			action:        "delete",
			expectedError: ErrForbidden,
		},
		{
			name: "maintainer-delete-allowed",
			user: User{
				Id:   3,
				Name: "user 3",
				Roles: []string{
					roleTicketMaintainer,
				},
			},
			ticket: Ticket{
				Id:   2,
				Body: "lorem ipsum",
				Type: "general",
			},
			action:        "delete",
			expectedError: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := o.Authorize(c.user, c.action, c.ticket)
			if c.expectedError == nil {
				if err == nil {
					return
				}
				t.Fatalf("expected: allow, got error: %v", err)
			} else {
				if err == nil {
					t.Fatalf("expected: reject, got nil error")
				} else if c.expectedError != err {
					t.Logf("got wrong error: expected=%v, got=%v", c.expectedError, err)
				}
			}

		})
	}
}
