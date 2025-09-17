package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const expectedParts = 2

func setupUsage() {
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [options]\n\n",
			os.Args[0],
		)

		fmt.Fprintf(
			os.Stderr,
			"Find Go interface implementations in a codebase.\n\n",
		)

		fmt.Fprintf(
			os.Stderr,
			"Options:\n",
		)

		flag.PrintDefaults()

		fmt.Fprintf(
			os.Stderr,
			"\nExample:\n",
		)

		fmt.Fprintf(
			os.Stderr,
			"  %s -interface ./internal/app/server.go:Server -dir ./internal/pkg/\n",
			os.Args[0],
		)
	}
}

func configureLogging(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
}

func runFinder(interfaceFile, interfaceName, searchDir string) error {
	if err := validateArgs(
		interfaceFile,
		interfaceName,
		searchDir,
	); err != nil {
		return err
	}

	finder := NewFinder(interfaceName)

	if err := finder.validateGoModRoot(); err != nil {
		return err
	}

	if err := finder.loadModulePath(); err != nil {
		return err
	}

	if err := finder.parseInterface(interfaceFile); err != nil {
		return err
	}

	logrus.WithField("methods", finder.interfaceMethods).
		Debugf("Found %d interface methods", len(finder.interfaceMethods))

	if err := finder.scanDirectory(searchDir); err != nil {
		return err
	}

	logrus.Debugf(
		"Scan complete, found %d implementations",
		len(finder.results),
	)

	implementations := finder.getResults()

	output, err := json.MarshalIndent(
		implementations,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf("failed to marshal implementations to JSON: %w", err)
	}

	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("failed to write output to stdout: %w", err)
	}

	if _, err := os.Stdout.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline to stdout: %w", err)
	}

	return nil
}

func parseInterfaceSpec(
	spec string,
) (string, string, error) {
	if spec == "" {
		return "", "", ErrInterfaceSpecEmpty
	}

	parts := strings.Split(spec, ":")
	if len(parts) != expectedParts {
		return "", "", ErrInterfaceSpecFormat
	}

	interfaceFile := strings.TrimSpace(parts[0])
	interfaceName := strings.TrimSpace(parts[1])

	if interfaceFile == "" {
		return "", "", ErrInterfaceFilePathEmpty
	}

	if interfaceName == "" {
		return "", "", ErrInterfaceNameEmpty
	}

	return interfaceFile, interfaceName, nil
}

func main() {
	var (
		interfaceSpec = flag.String(
			"interface",
			"",
			"Interface specification in format 'file.go:InterfaceName'",
		)

		searchDir = flag.String(
			"dir",
			".",
			"Directory to search for implementations",
		)

		help = flag.Bool(
			"help",
			false,
			"Show help",
		)

		debug = flag.Bool(
			"debug",
			false,
			"Enable debug logging",
		)
	)

	setupUsage()
	flag.Parse()

	configureLogging(*debug)

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	interfaceFile, interfaceName, err := parseInterfaceSpec(
		*interfaceSpec,
	)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.WithFields(logrus.Fields{
		"interface_file": interfaceFile,
		"interface_name": interfaceName,
		"search_dir":     *searchDir,
	}).Debug("Parsed arguments")

	if err := runFinder(interfaceFile, interfaceName, *searchDir); err != nil {
		logrus.Fatal(err)
	}
}
