package oso

import (
	"errors"
	"fmt"

	"github.com/osohq/go-oso"
	"github.com/sirupsen/logrus"
	"github.com/umutozd/iam-prototypes/auth"
)

var ErrNotFound = errors.New("oso: user or access rule not found")
var ErrForbidden = errors.New("oso: user is forbidden to perform this action")

type osoAuth struct {
	o *oso.Oso
}

func NewOsoAuth(registeredClasses []any, policyString string) (auth.AuthLib, error) {
	o, err := oso.NewOso()
	if err != nil {
		return nil, fmt.Errorf("oso: %v", err)
	}

	o.SetReadAction("read")
	o.SetNotFoundError(func() error { return ErrNotFound })
	o.SetForbiddenError(func() error { return ErrForbidden })

	registeredClasses = append(registeredClasses, User{})
	for _, class := range registeredClasses {
		if err = o.RegisterClass(class, nil); err != nil {
			return nil, fmt.Errorf("oso: register class error: %v, class: %#v", err, class)
		}
	}

	if err = o.LoadString(policyString); err != nil {
		return nil, fmt.Errorf("oso: load policy error: %v", err)
	}

	return &osoAuth{
		o: &o,
	}, nil
}

func (o *osoAuth) Authorize(userId string, usergroups []string, resource any, action string) bool {
	logrus.WithFields(logrus.Fields{
		"user_id":  userId,
		"groups":   usergroups,
		"resource": resource,
		"action":   action,
	}).Infof("oso: authorizing")
	if resource == nil {
		logrus.Warn("oso: resource is nil")
	}
	err := o.o.Authorize(
		User{Id: userId, Roles: usergroups},
		action,
		resource,
	)
	return err == nil
}

func (o *osoAuth) Name() auth.AuthLibName {
	return auth.AuthLibNameOso
}

// User is a struct to be used by the internal oso library.
type User struct {
	Id    string
	Roles []string
}
