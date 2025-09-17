package piraterf

import (
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gonfiguration"
)

const (
	envVarNameHTMLDir          = "PIRATERF_HTMLDIR"
	envVarNameStaticDir        = "PIRATERF_STATICDIR"
	envVarNamePiraterfFilesDir = "PIRATERF_FILESDIR"
	envVarNameUploadDir        = "PIRATERF_UPLOADDIR"

	defaultHTMLDir   = "./html"
	defaultStaticDir = "./static"
	defaultFilesDir  = "./files"
	defaultUploadDir = "./uploads"
)

type Config struct {
	HTMLDir   string `env:"PIRATERF_HTMLDIR"`
	StaticDir string `env:"PIRATERF_STATICDIR"`
	FilesDir  string `env:"PIRATERF_FILESDIR"`
	UploadDir string `env:"PIRATERF_UPLOADDIR"`
}

func parseConfig() (Config, error) {
	cfg := Config{}

	gonfiguration.SetDefaults(map[string]any{
		envVarNamePiraterfFilesDir: defaultFilesDir,
		envVarNameHTMLDir:          defaultHTMLDir,
		envVarNameStaticDir:        defaultStaticDir,
		envVarNameUploadDir:        defaultUploadDir,
	})

	if err := gonfiguration.Parse(&cfg); err != nil {
		return Config{}, ctxerrors.Wrap(err, "could not parse config")
	}

	return cfg, nil
}
