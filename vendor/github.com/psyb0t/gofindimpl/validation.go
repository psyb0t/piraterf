package main

import (
	"fmt"
	"os"
)

func validateArgs(
	interfaceFile, _ string, searchDir string,
) error {
	if _, err := os.Stat(interfaceFile); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrInterfaceFileNotExist, interfaceFile)
	}

	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrSearchDirNotExist, searchDir)
	}

	return nil
}
