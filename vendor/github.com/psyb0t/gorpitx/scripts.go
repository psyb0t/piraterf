package gorpitx

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/psyb0t/ctxerrors"
)

// fskScript contains the embedded FSK script content
//
//go:embed scripts/fsk.sh
var fskScript string

// ModuleNameToScriptName returns the script path for script-based modules.
func ModuleNameToScriptName(moduleName ModuleName) (string, bool) {
	switch moduleName {
	case ModuleNameFSK:
		return "/tmp/fsk.sh", true
	default:
		return "", false
	}
}

// EnsureScriptExists writes the embedded script to filesystem if it doesn't
// exist.
func EnsureScriptExists(moduleName ModuleName) error {
	scriptPath, isScript := ModuleNameToScriptName(moduleName)
	if !isScript {
		return nil // Not a script-based module
	}

	// Check if script already exists
	if _, err := os.Stat(scriptPath); err == nil {
		return nil // Script already exists
	}

	var scriptContent string

	switch moduleName {
	case ModuleNameFSK:
		scriptContent = fskScript
	default:
		return ctxerrors.Wrapf(
			ErrUnknownModule,
			"no script content for module: %s",
			moduleName,
		)
	}

	const (
		dirPerm    = 0o750
		scriptPerm = 0o600
		execPerm   = 0o700
	)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(scriptPath), dirPerm); err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to create script directory: %s",
			filepath.Dir(scriptPath),
		)
	}

	// Write script to filesystem
	if err := os.WriteFile(
		scriptPath, []byte(scriptContent), scriptPerm,
	); err != nil {
		return ctxerrors.Wrapf(err, "failed to write script: %s", scriptPath)
	}

	// Make script executable
	if err := os.Chmod(scriptPath, execPerm); err != nil {
		return ctxerrors.Wrapf(
			err,
			"failed to make script executable: %s",
			scriptPath,
		)
	}

	return nil
}

// IsScriptModule returns true if the module uses an embedded script.
func IsScriptModule(moduleName ModuleName) bool {
	_, isScript := ModuleNameToScriptName(moduleName)

	return isScript
}
