package service

import "github.com/umutozd/iam-prototypes/auth"

type Config struct {
	Debug   bool
	Port    int
	AuthLib auth.AuthLibName
}
