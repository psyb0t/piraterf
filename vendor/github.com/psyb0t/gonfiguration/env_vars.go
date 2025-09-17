package gonfiguration

import (
	"maps"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	envVarNumParts = 2
)

func GetEnvVars() map[string]string {
	return gonfig.getEnvVars()
}

func (g *gonfiguration) setEnvVar(key string, val string) {
	g.Lock()
	defer g.Unlock()

	g.envVars[key] = val
}

func (g *gonfiguration) setEnvVars(envVars map[string]string) {
	for key, val := range envVars {
		g.setEnvVar(key, val)
	}
}

func (g *gonfiguration) getEnvVars() map[string]string {
	g.RLock()
	defer g.RUnlock()

	envVarsCopy := make(map[string]string, len(g.envVars))
	maps.Copy(envVarsCopy, g.envVars)

	return envVarsCopy
}

func getEnvVars() (map[string]string, error) {
	envVars := map[string]string{}
	rawVars := os.Environ()

	for _, rawVar := range rawVars {
		parts := strings.SplitN(rawVar, "=", envVarNumParts)
		if len(parts) != envVarNumParts {
			return nil, errors.Wrap(ErrInvalidEnvVar, rawVar)
		}

		envVars[parts[0]] = parts[1]
	}

	return envVars, nil
}
