package main

import "errors"

var (
	ErrGoModNotFound = errors.New(
		"go.mod not found in current directory - must launch from go module root")
	ErrNoModuleDeclaration   = errors.New("no module declaration found in go.mod")
	ErrInterfaceFileRequired = errors.New(
		"interface file is required. Use -interface flag")
	ErrInterfaceNameRequired = errors.New(
		"interface name is required. Use -name flag")
	ErrInterfaceNotFound  = errors.New("interface not found in file")
	ErrNoFilesToTypeCheck = errors.New("no files to type check")
	ErrInterfaceSpecEmpty = errors.New(
		"interface specification cannot be empty. " +
			"Use -interface flag with format 'file.go:InterfaceName'")
	ErrInterfaceSpecFormat = errors.New(
		"interface specification must be in format 'file.go:InterfaceName'")
	ErrInterfaceFilePathEmpty = errors.New("interface file path cannot be empty")
	ErrInterfaceNameEmpty     = errors.New("interface name cannot be empty")
	ErrInterfaceFileNotExist  = errors.New("interface file does not exist")
	ErrSearchDirNotExist      = errors.New("search directory does not exist")
)
