package main

import (
	"github.com/sirupsen/logrus"
	"github.com/umutozd/iam-prototypes/auth"
	"github.com/umutozd/iam-prototypes/service"
	"github.com/urfave/cli/v2"
)

var config = &service.Config{
	Port:    8080,
	AuthLib: auth.AuthLibNameCasbin,
}

func main() {
	app := cli.NewApp()
	app.Name = "iam-prototypes"
	app.Usage = "A gRPC server that showcases use of several authorization libraries."
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Destination: &config.Debug,
			Usage:       "enables verbose logging",
			Aliases:     []string{"d"},
		},
		&cli.IntFlag{
			Name:        "port",
			Value:       config.Port,
			Destination: &config.Port,
			Usage:       "the port to listen to gRPC connections",
			Aliases:     []string{"p"},
		},
		&cli.StringFlag{
			Name:        "lib",
			Value:       string(config.AuthLib),
			Destination: (*string)(&config.AuthLib),
			Usage:       "the authorization library to use",
		},
	}

	srv, err := service.NewServer(config)
	if err != nil {
		logrus.WithError(err).Fatal("error creating server")
	}
	if err = srv.Listen(); err != nil {
		logrus.WithError(err).Fatal("server stopped with error")
	}
}
