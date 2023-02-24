package main

import (
	"fmt"
	"os"

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

	app.Action = runAction
	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("server exited with error")
	}
}

func runAction(c *cli.Context) error {
	srv, err := service.NewServer(config)
	if err != nil {
		return fmt.Errorf("error creating server: %v", err)
	}
	return srv.Listen()
}
