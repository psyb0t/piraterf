package server

import (
	"time"

	"github.com/psyb0t/aichteeteapee"
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gonfiguration"
)

type Config struct {
	ListenAddress       string        `env:"HTTP_SERVER_LISTENADDRESS"`
	ReadTimeout         time.Duration `env:"HTTP_SERVER_READTIMEOUT"`
	ReadHeaderTimeout   time.Duration `env:"HTTP_SERVER_READHEADERTIMEOUT"`
	WriteTimeout        time.Duration `env:"HTTP_SERVER_WRITETIMEOUT"`
	IdleTimeout         time.Duration `env:"HTTP_SERVER_IDLETIMEOUT"`
	MaxHeaderBytes      int           `env:"HTTP_SERVER_MAXHEADERBYTES"`
	ShutdownTimeout     time.Duration `env:"HTTP_SERVER_SHUTDOWNTIMEOUT"`
	ServiceName         string        `env:"HTTP_SERVER_SERVICENAME"`
	FileUploadMaxMemory int64         `env:"HTTP_SERVER_FILEUPLOADMAXMEMORY"`
	TLSEnabled          bool          `env:"HTTP_SERVER_TLSENABLED"`
	TLSListenAddress    string        `env:"HTTP_SERVER_TLSLISTENADDRESS"`
	TLSCertFile         string        `env:"HTTP_SERVER_TLSCERTFILE"`
	TLSKeyFile          string        `env:"HTTP_SERVER_TLSKEYFILE"`
}

// parseConfig parses the server configuration from environment variables.
func parseConfig() (Config, error) {
	cfg := Config{}

	//nolint:lll
	gonfiguration.SetDefaults(map[string]any{
		aichteeteapee.EnvVarNameHTTPServerListenAddress:       aichteeteapee.DefaultHTTPServerListenAddress,
		aichteeteapee.EnvVarNameHTTPServerReadTimeout:         aichteeteapee.DefaultHTTPServerReadTimeout,
		aichteeteapee.EnvVarNameHTTPServerReadHeaderTimeout:   aichteeteapee.DefaultHTTPServerReadHeaderTimeout,
		aichteeteapee.EnvVarNameHTTPServerWriteTimeout:        aichteeteapee.DefaultHTTPServerWriteTimeout,
		aichteeteapee.EnvVarNameHTTPServerIdleTimeout:         aichteeteapee.DefaultHTTPServerIdleTimeout,
		aichteeteapee.EnvVarNameHTTPServerMaxHeaderBytes:      aichteeteapee.DefaultHTTPServerMaxHeaderBytes,
		aichteeteapee.EnvVarNameHTTPServerShutdownTimeout:     aichteeteapee.DefaultHTTPServerShutdownTimeout,
		aichteeteapee.EnvVarNameHTTPServerServiceName:         aichteeteapee.DefaultHTTPServerServiceName,
		aichteeteapee.EnvVarNameHTTPServerFileUploadMaxMemory: aichteeteapee.DefaultFileUploadMaxMemory,
		aichteeteapee.EnvVarNameHTTPServerTLSEnabled:          aichteeteapee.DefaultHTTPServerTLSEnabled,
		aichteeteapee.EnvVarNameHTTPServerTLSListenAddress:    aichteeteapee.DefaultHTTPServerTLSListenAddress,
		aichteeteapee.EnvVarNameHTTPServerTLSCertFile:         aichteeteapee.DefaultHTTPServerTLSCertFile,
		aichteeteapee.EnvVarNameHTTPServerTLSKeyFile:          aichteeteapee.DefaultHTTPServerTLSKeyFile,
	})

	if err := gonfiguration.Parse(&cfg); err != nil {
		return Config{}, ctxerrors.Wrap(err, "could not parse config")
	}

	return cfg, nil
}
