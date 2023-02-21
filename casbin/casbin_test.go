package casbin

import (
	"os"
	"testing"

	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"
)

func TestCasbin(t *testing.T) {
	e, err := casbin.NewEnforcer("model.conf", "policy.csv")
	if err != nil {
		t.Fatalf("error creating new enforcer: %v", err)
	}

	runCasbinCases(t, e)
}

func TestCasbinWithPostgresAdapter(t *testing.T) {
	// create adapter and enforcer
	dbUrl := os.Getenv("TEST_CASBIN_POSTGRESQL_URL")
	a, err := pgadapter.NewAdapter(dbUrl, "postgres")
	if err != nil {
		t.Fatalf("error creating postgres adapter: %v", err)
	}
	e, err := casbin.NewEnforcer("model.conf", a)
	if err != nil {
		t.Fatalf("error creating casbin enforcer with pg adapter: %v", err)
	}

	// add policies that are in "policy.csv"
	_, err = e.AddNamedGroupingPoliciesEx("g", [][]string{
		{"otsimo.com/ticket/editor", "otsimo.com/ticket/viewer"},
		{"otsimo.com/ticket/maintainer", "otsimo.com/ticket/editor"},

		{"selahattin", "otsimo.com/ticket/viewer"},
		{"umut", "otsimo.com/ticket/editor"},
		{"sercan", "otsimo.com/ticket/maintainer"},
	})
	if err != nil {
		t.Fatalf("error adding role policies: %v", err)
	}

	_, err = e.AddNamedPoliciesEx("p", [][]string{
		{"otsimo.com/ticket/viewer", "ticket", "read"},
		{"otsimo.com/ticket/editor", "ticket", "write"},
		{"otsimo.com/ticket/maintainer", "ticket", "delete"},
		{"elif", "ticket", "write"},
	})
	if err != nil {
		t.Fatalf("error adding policies: %v", err)
	}
	if err = e.SavePolicy(); err != nil {
		t.Fatalf("error saving policy to db: %v", err)
	}

	runCasbinCases(t, e)
}

func runCasbinCases(t *testing.T, e *casbin.Enforcer) {
	cases := []*struct {
		name           string
		subj           string
		obj            string
		act            string
		expectedResult bool
	}{
		{
			name:           "umut-write-allow",
			subj:           "umut",
			obj:            "ticket",
			act:            "write",
			expectedResult: true,
		},
		{
			name:           "umut-delete-forbid",
			subj:           "umut",
			obj:            "ticket",
			act:            "delete",
			expectedResult: false,
		},
		{
			name:           "umut-read-allow",
			subj:           "umut",
			obj:            "ticket",
			act:            "read",
			expectedResult: true,
		},
		{
			name:           "sercan-read-allow",
			subj:           "sercan",
			obj:            "ticket",
			act:            "read",
			expectedResult: true,
		},
		{
			name:           "sercan-write-allow",
			subj:           "sercan",
			obj:            "ticket",
			act:            "write",
			expectedResult: true,
		},
		{
			name:           "sercan-delete-allow",
			subj:           "sercan",
			obj:            "ticket",
			act:            "delete",
			expectedResult: true,
		},
		{
			name:           "selahattin-delete-forbid",
			subj:           "selahattin",
			obj:            "ticket",
			act:            "delete",
			expectedResult: false,
		},
		{
			name:           "selahattin-write-forbid",
			subj:           "selahattin",
			obj:            "ticket",
			act:            "write",
			expectedResult: false,
		},
		{
			name:           "selahattin-read-allow",
			subj:           "selahattin",
			obj:            "ticket",
			act:            "read",
			expectedResult: true,
		},
		{
			name:           "non-existing-subject-forbid",
			subj:           "begüm",
			obj:            "ticket",
			act:            "read",
			expectedResult: false,
		},
		{
			name:           "non-existing-object-forbid",
			subj:           "sercan",
			obj:            "trigger",
			act:            "read",
			expectedResult: false,
		},
		{
			name:           "non-existing-action-forbid",
			subj:           "sercan",
			obj:            "ticket",
			act:            "own",
			expectedResult: false,
		},
		{
			name:           "access-through-non-rbac",
			subj:           "elif",
			obj:            "ticket",
			act:            "write",
			expectedResult: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			allowed, err := e.Enforce(c.subj, c.obj, c.act)
			if err != nil {
				t.Fatalf("Enforce returned error: %v", err)
			}
			if allowed != c.expectedResult {
				t.Fatalf("wrong result: expected=%t, got=%t", c.expectedResult, allowed)
			}
		})
	}
}
