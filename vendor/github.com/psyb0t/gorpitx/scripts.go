package gorpitx

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/psyb0t/ctxerrors"
	"github.com/sirupsen/logrus"
)

const (
	fskScriptPath          = "/tmp/fsk.sh"
	audioSockBroadcastPath = "/tmp/audiosock_broadcast.sh"
	modulationPath         = "/tmp/modulation.sh"

	dirPerm    = 0o750
	scriptPerm = 0o600
	execPerm   = 0o700
)

// fskScript contains the embedded FSK script content
//
//go:embed scripts/fsk.sh
var fskScript string

// audioSockBroadcastScript contains the embedded AudioSock script
//
//go:embed scripts/audiosock_broadcast.sh
var audioSockBroadcastScript string

// modulationScript contains the embedded modulation script
//
//go:embed scripts/modulation.sh
var modulationScript string

// init writes all embedded scripts to filesystem on package initialization.
//
//nolint:gochecknoinits // Required for automatic script deployment
func init() {
	writeAllScripts()
}

// writeAllScripts writes all embedded scripts to filesystem unconditionally.
//
//nolint:funlen // Function length due to proper parameter formatting
func writeAllScripts() {
	var err error

	// Create directories
	err = os.MkdirAll(
		filepath.Dir(fskScriptPath),
		dirPerm,
	)
	if err != nil {
		logrus.Fatalf("failed to create script directory: %v", err)
	}

	err = os.MkdirAll(
		filepath.Dir(audioSockBroadcastPath),
		dirPerm,
	)
	if err != nil {
		logrus.Fatalf("failed to create script directory: %v", err)
	}

	err = os.MkdirAll(
		filepath.Dir(modulationPath),
		dirPerm,
	)
	if err != nil {
		logrus.Fatalf("failed to create script directory: %v", err)
	}

	// Write FSK script
	err = os.WriteFile(
		fskScriptPath,
		[]byte(fskScript),
		scriptPerm,
	)
	if err != nil {
		logrus.Fatalf("failed to write FSK script: %v", err)
	}

	err = os.Chmod(fskScriptPath, execPerm)
	if err != nil {
		logrus.Fatalf("failed to make FSK script executable: %v", err)
	}

	// Write AudioSock script
	err = os.WriteFile(
		audioSockBroadcastPath,
		[]byte(audioSockBroadcastScript),
		scriptPerm,
	)
	if err != nil {
		logrus.Fatalf("failed to write AudioSock script: %v", err)
	}

	err = os.Chmod(audioSockBroadcastPath, execPerm)
	if err != nil {
		logrus.Fatalf("failed to make AudioSock script executable: %v", err)
	}

	// Write modulation script
	err = os.WriteFile(
		modulationPath,
		[]byte(modulationScript),
		scriptPerm,
	)
	if err != nil {
		logrus.Fatalf("failed to write modulation script: %v", err)
	}

	err = os.Chmod(modulationPath, execPerm)
	if err != nil {
		logrus.Fatalf("failed to make modulation script executable: %v", err)
	}
}

// ModuleNameToScriptName returns the script path for script-based modules.
func ModuleNameToScriptName(moduleName ModuleName) (string, bool) {
	switch moduleName {
	case ModuleNameFSK:
		return fskScriptPath, true
	case ModuleNameAudioSockBroadcast:
		return audioSockBroadcastPath, true
	default:
		return "", false
	}
}

// EnsureScriptExists writes the embedded script if it doesn't exist.
func EnsureScriptExists(moduleName ModuleName) error {
	scriptPath, isScript := ModuleNameToScriptName(moduleName)
	if !isScript {
		return nil
	}

	if scriptExists(scriptPath) {
		return ensureAudioSockModulation(moduleName)
	}

	return writeScript(moduleName, scriptPath)
}

// scriptExists checks if a script file exists.
func scriptExists(scriptPath string) bool {
	_, err := os.Stat(scriptPath)

	return err == nil
}

// ensureAudioSockModulation ensures modulation script exists for AudioSock.
func ensureAudioSockModulation(moduleName ModuleName) error {
	if moduleName != ModuleNameAudioSockBroadcast {
		return nil
	}

	if _, err := os.Stat(modulationPath); err != nil {
		return ensureModulationScript(scriptPerm, execPerm)
	}

	return nil
}

// writeScript writes a script to the filesystem.
func writeScript(moduleName ModuleName, scriptPath string) error {
	scriptContent, err := getScriptContent(moduleName)
	if err != nil {
		return err
	}

	if err := createScriptDir(scriptPath); err != nil {
		return err
	}

	if err := writeScriptFile(scriptPath, scriptContent); err != nil {
		return err
	}

	if err := makeExecutable(scriptPath); err != nil {
		return err
	}

	return ensureAudioSockModulation(moduleName)
}

// getScriptContent returns the embedded script content for a module.
func getScriptContent(moduleName ModuleName) (string, error) {
	switch moduleName {
	case ModuleNameFSK:
		return fskScript, nil
	case ModuleNameAudioSockBroadcast:
		return audioSockBroadcastScript, nil
	default:
		return "", ctxerrors.Wrapf(
			ErrUnknownModule,
			"no script content for module: %s",
			moduleName,
		)
	}
}

// createScriptDir creates the script directory if it doesn't exist.
func createScriptDir(scriptPath string) error {
	err := os.MkdirAll(
		filepath.Dir(scriptPath),
		dirPerm,
	)
	if err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to create script directory: %s",
			filepath.Dir(scriptPath),
		)
	}

	return nil
}

// writeScriptFile writes the script content to a file.
func writeScriptFile(scriptPath, content string) error {
	err := os.WriteFile(
		scriptPath,
		[]byte(content),
		scriptPerm,
	)
	if err != nil {
		return ctxerrors.Wrapf(err, "failed to write script: %s", scriptPath)
	}

	return nil
}

// makeExecutable makes a script file executable.
func makeExecutable(scriptPath string) error {
	err := os.Chmod(scriptPath, execPerm)
	if err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to make script executable: %s",
			scriptPath,
		)
	}

	return nil
}

// ensureModulationScript writes modulation.sh if it doesn't exist.
func ensureModulationScript(scriptPerm, execPerm os.FileMode) error {
	// Check if script already exists
	if _, err := os.Stat(modulationPath); err == nil {
		return nil // Script already exists
	}

	if err := os.WriteFile(
		modulationPath,
		[]byte(modulationScript),
		scriptPerm,
	); err != nil {
		return ctxerrors.Wrapf(err,
			"failed to write modulation.sh: %s", modulationPath)
	}

	// Make modulation.sh executable
	if err := os.Chmod(modulationPath, execPerm); err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to make modulation.sh executable: %s",
			modulationPath,
		)
	}

	return nil
}

// IsScriptModule returns true if the module uses an embedded script.
func IsScriptModule(moduleName ModuleName) bool {
	_, isScript := ModuleNameToScriptName(moduleName)

	return isScript
}
