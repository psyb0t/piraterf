package gorpitx

import (
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gonfiguration"
)

const (
	envVarNameGorpitxPath = "GORPITX_PATH"
	defaultPath           = "$HOME/rpitx"
)

type Config struct {
	Path string `env:"GORPITX_PATH"`
}

func parseConfig() (Config, error) {
	cfg := Config{}

	gonfiguration.SetDefaults(map[string]any{
		envVarNameGorpitxPath: defaultPath,
	})

	if err := gonfiguration.Parse(&cfg); err != nil {
		return Config{}, ctxerrors.Wrap(err, "could not parse config")
	}

	return cfg, nil
}
