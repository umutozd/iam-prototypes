package service

import (
	"fmt"

	"github.com/sercand/graceful"
	"github.com/umutozd/iam-prototypes/auth"
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
	switch cfg.AuthLib {
	case auth.AuthLibNameCasbin:
		// TODO
		// s.auth, err = casbin.NewCasbinAuth()
	case auth.AuthLibNameGorbac:
		// TODO
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
