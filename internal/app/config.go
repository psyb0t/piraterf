package app

import (
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gonfiguration"
)

// Ready for your gorgeous ass config vals.
type config struct {
	NiceField string `env:"APP_NICEFIELD"`
}

func parseConfig() (config, error) {
	cfg := config{}

	gonfiguration.SetDefaults(map[string]any{})

	if err := gonfiguration.Parse(&cfg); err != nil {
		return config{}, ctxerrors.Wrap(err, "could not parse config")
	}

	return cfg, nil
}
