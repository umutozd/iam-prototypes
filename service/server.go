package service

import (
	"fmt"

	"github.com/sercand/graceful"
	"github.com/sirupsen/logrus"
	"github.com/umutozd/iam-prototypes/auth"
	"github.com/umutozd/iam-prototypes/auth/casbin"
	"github.com/umutozd/iam-prototypes/auth/gorbac"
	"github.com/umutozd/iam-prototypes/auth/oso"
	"github.com/umutozd/iam-prototypes/pb"
	"google.golang.org/grpc"
)

type Server struct {
	cfg  *Config
	auth auth.AuthLib
}

func NewServer(cfg *Config) (*Server, error) {
	var err error
	s := &Server{
		cfg: cfg,
	}
	logrus.Infof("using auth-lib %q", cfg.AuthLib)
	switch cfg.AuthLib {
	case auth.AuthLibNameCasbin:
		s.auth, err = casbin.NewCasbinAuth(pb.SimpleServiceAuthConfig.CasbinConfig.Permissions, pb.SimpleServiceAuthConfig.CasbinConfig.InheritanceRules)
	case auth.AuthLibNameGorbac:
		cfg := pb.SimpleServiceAuthConfig.GorbacConfig
		s.auth, err = gorbac.NewGorbacAuth(cfg.Roles, cfg.Permissions, cfg.PermissionsToRoles, cfg.InheritanceRules)
	case auth.AuthLibNameOso:
		s.auth, err = oso.NewOsoAuth([]any{pb.Foo{}}, pb.OsoPolicy)
	default:
		err = fmt.Errorf("invalid AuthLib: %s", cfg.AuthLib)
	}
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Listen() error {
	g := grpc.NewServer(
		grpc.UnaryInterceptor(
			s.AuthInterceptor(pb.SimpleServiceAuthConfig),
		),
	)
	pb.RegisterSimpleServiceServer(g, s)

	return <-graceful.ServeAndStopOnSignals(
		graceful.NewServable(fmt.Sprintf(":%d", s.cfg.Port), g),
	)
}
