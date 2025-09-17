package env

import (
	"os"
)

const (
	EnvVarName string = "ENV"
)

type Type = string

const (
	EnvTypeProd Type = "prod"
	EnvTypeDev  Type = "dev"
)

func Get() Type {
	e := os.Getenv(EnvVarName)

	switch e {
	case EnvTypeDev:
		return EnvTypeDev
	default:
		return EnvTypeProd
	}
}

func IsProd() bool {
	return Get() == EnvTypeProd
}

func IsDev() bool {
	return Get() == EnvTypeDev
}
