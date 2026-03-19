package goenv

import (
	"os"
)

const EnvVarName = "ENV"

type Type = string

const (
	Prod Type = "prod"
	Dev  Type = "dev"
)

func Get() Type {
	e := os.Getenv(EnvVarName)

	switch e {
	case Dev:
		return Dev
	default:
		return Prod
	}
}

func IsProd() bool {
	return Get() == Prod
}

func IsDev() bool {
	return Get() == Dev
}
