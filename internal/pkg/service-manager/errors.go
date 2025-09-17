package servicemanager

import "errors"

var (
	ErrServiceNotFound   = errors.New("service not found")
	ErrNoEnabledServices = errors.New("no enabled services")
)
