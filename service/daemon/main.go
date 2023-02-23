package main

import (
	"github.com/sirupsen/logrus"
	"github.com/umutozd/iam-prototypes/auth"
	"github.com/umutozd/iam-prototypes/service"
)

func main() {
	cfg := &service.Config{
		Port:    8080,
		AuthLib: auth.AuthLibNameOso,
	}

	srv, err := service.NewServer(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("error creating server")
	}
	if err = srv.Listen(); err != nil {
		logrus.WithError(err).Fatal("server stopped with error")
	}
}