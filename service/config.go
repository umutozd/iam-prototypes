package service

import "github.com/umutozd/iam-prototypes/auth"

type Config struct {
	Port    int
	AuthLib auth.AuthLibName
}
