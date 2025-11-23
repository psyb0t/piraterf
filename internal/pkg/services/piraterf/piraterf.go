package piraterf

import (
	"bytes"
	"context"
	"os"
	"path"
	"sync"
	"text/template"

	"github.com/psyb0t/aichteeteapee/server"
	"github.com/psyb0t/aichteeteapee/server/dabluvee-es/wshub"
	"github.com/psyb0t/commander"
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gorpitx"
	"github.com/sirupsen/logrus"
)

const (
	ServiceName       = "PIrateRF"
	audioFilesDir     = "audio"
	uploadsSubdir     = "uploads"
	audioSFXDir       = "sfx"
	audioUploadsPath  = audioFilesDir + "/" + uploadsSubdir
	imagesFilesDir    = "images"
	imagesUploadsPath = imagesFilesDir + "/" + uploadsSubdir
	dataFilesDir      = "data"
	dataUploadsPath   = dataFilesDir + "/" + uploadsSubdir
	iqsFilesDir       = "iqs"
	iqsUploadsPath    = iqsFilesDir + "/" + uploadsSubdir
	presetsDir        = "presets"
	envJSFilename     = "env.js"
	envJSTemplate     = `window.PIrateRFConfig = {
  paths: {
    files: "/files",
    audioUploadFiles: "/files/` + audioFilesDir + `/` + uploadsSubdir + `",
    audioSFXFiles: "/files/` + audioFilesDir + `/` + audioSFXDir + `",
    imageUploadFiles: "/files/` + imagesFilesDir + `/` + uploadsSubdir + `",
    dataUploadFiles: "/files/` + dataFilesDir + `/` + uploadsSubdir + `",
    iqUploadFiles: "/files/` + iqsFilesDir + `/` + uploadsSubdir + `",
    presets: "/files/` + presetsDir + `"
  },
  directories: {
    audioFiles: "` + audioFilesDir + `",
    audioUploads: "` + audioFilesDir + `/` + uploadsSubdir + `",
    audioSFX: "` + audioFilesDir + `/` + audioSFXDir + `",
    imageFiles: "` + imagesFilesDir + `",
    imageUploads: "` + imagesFilesDir + `/` + uploadsSubdir + `",
    dataFiles: "` + dataFilesDir + `",
    dataUploads: "` + dataFilesDir + `/` + uploadsSubdir + `",
    iqFiles: "` + iqsFilesDir + `",
    iqUploads: "` + iqsFilesDir + `/` + uploadsSubdir + `"
  },
  serverPaths: {
    audioUploads: "{{.FilesDir}}/` + audioFilesDir + `/` +
		uploadsSubdir + `",
    audioSFX: "{{.FilesDir}}/` + audioFilesDir + `/` + audioSFXDir + `",
    imageUploads: "{{.FilesDir}}/` + imagesFilesDir + `/` +
		uploadsSubdir + `",
    dataUploads: "{{.FilesDir}}/` + dataFilesDir + `/` +
		uploadsSubdir + `",
    iqUploads: "{{.FilesDir}}/` + iqsFilesDir + `/` +
		uploadsSubdir + `"
  }
};
`
	// Audio format constants for sox conversion.
	audioSampleRate = "48000" // 48kHz sample rate
	audioBitDepth   = "16"    // 16-bit depth
	audioChannels   = "1"     // mono (1 channel)

	// File and directory permissions.
	// readable/executable by all for web serving.
	dirPerms  = 0o755
	filePerms = 0o644 // readable by all for web serving
)

type PIrateRF struct {
	config           Config
	rpitx            *gorpitx.RPITX
	httpServer       *server.Server
	websocketHub     wshub.Hub
	executionManager *executionManager
	commander        commander.Commander
	serviceCtx       context.Context //nolint:containedctx
	// need service ctx to pass down to process execution
	doneCh   chan struct{}
	stopOnce sync.Once
}

func New() (*PIrateRF, error) {
	cfg, err := parseConfig()
	if err != nil {
		return nil, ctxerrors.Wrap(err, "could not parse config")
	}

	return NewWithConfig(cfg)
}

func NewWithConfig(config Config) (*PIrateRF, error) {
	s := &PIrateRF{
		config:    config,
		rpitx:     gorpitx.GetInstance(),
		commander: commander.New(),
		doneCh:    make(chan struct{}),
	}

	// Ensure directories exist during construction
	if err := s.ensureUploadDirExists(); err != nil {
		return nil, ctxerrors.Wrap(err, "failed to ensure upload directory exists")
	}

	if err := s.ensureFilesDirsExist(); err != nil {
		return nil, ctxerrors.Wrap(err, "failed to ensure files directories exist")
	}

	// Generate env.js config file for frontend
	if err := s.generateEnvJS(); err != nil {
		return nil, ctxerrors.Wrap(err, "failed to generate env.js config")
	}

	s.setupWebsocketHub()
	s.executionManager = newExecutionManager(s.rpitx, s.websocketHub)

	if err := s.setupHTTPServer(); err != nil {
		return nil, ctxerrors.Wrap(err, "could not setup http server")
	}

	return s, nil
}

func (s *PIrateRF) Name() string {
	return ServiceName
}

func (s *PIrateRF) Run(ctx context.Context) error {
	logrus.Infof("running %s service", ServiceName)

	// Store service context for event handlers
	s.serviceCtx = ctx

	defer func() {
		if err := s.Stop(ctx); err != nil {
			logrus.Errorf("failed to stop %s service: %v", ServiceName, err)
		}
	}()

	router, err := s.getHTTPServerRouter()
	if err != nil {
		return ctxerrors.Wrap(err, "failed to get HTTP server router")
	}

	httpServerErrCh := make(chan error, 1)

	go func() {
		defer close(httpServerErrCh)

		httpServerErrCh <- s.httpServer.Start(ctx, router)
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-s.doneCh:
		return nil
	case err := <-httpServerErrCh:
		if err != nil {
			return ctxerrors.Wrap(err, "http server error")
		}
	}

	return nil
}

func (s *PIrateRF) Stop(ctx context.Context) error {
	s.stopOnce.Do(func() {
		logrus.Infof("stopping %s service", ServiceName)

		close(s.doneCh)

		if err := s.httpServer.Stop(ctx); err != nil {
			logrus.Errorf("failed to stop http server: %v", err)
		}

		s.websocketHub.Close()
	})

	return nil
}

// ensureUploadDirExists creates the upload directory if it doesn't exist.
func (s *PIrateRF) ensureUploadDirExists() error {
	if err := os.MkdirAll(s.config.UploadDir, dirPerms); err != nil {
		return ctxerrors.Wrap(err, "failed to create upload directory")
	}

	return nil
}

// ensureFilesDirsExist creates the files directory structure if it doesn't
// exist.
func (s *PIrateRF) ensureFilesDirsExist() error {
	dirs := []struct {
		pathParts []string
		name      string
	}{
		{[]string{}, "files directory"},
		{[]string{audioFilesDir}, "audio directory"},
		{[]string{audioFilesDir, uploadsSubdir}, "audio uploads directory"},
		{[]string{audioFilesDir, audioSFXDir}, "audio SFX directory"},
		{[]string{imagesFilesDir}, "images directory"},
		{[]string{imagesFilesDir, uploadsSubdir}, "images uploads directory"},
		{[]string{dataFilesDir}, "data directory"},
		{[]string{dataUploadsPath}, "data uploads directory"},
		{[]string{iqsFilesDir}, "IQ directory"},
		{[]string{iqsUploadsPath}, "IQ uploads directory"},
		{[]string{presetsDir}, "presets directory"},
	}

	for _, dir := range dirs {
		parts := append([]string{s.config.FilesDir}, dir.pathParts...)
		dirPath := path.Join(parts...)

		if err := os.MkdirAll(dirPath, dirPerms); err != nil {
			return ctxerrors.Wrap(err, "failed to create "+dir.name)
		}
	}

	// Create preset subdirectories for all supported modules
	moduleNames := s.rpitx.GetSupportedModules()
	for _, moduleName := range moduleNames {
		modulePresetDir := path.Join(s.config.FilesDir, presetsDir, moduleName)
		if err := os.MkdirAll(modulePresetDir, dirPerms); err != nil {
			return ctxerrors.Wrap(
				err,
				"failed to create preset directory for module "+moduleName,
			)
		}
	}

	return nil
}

// generateEnvJS creates the env.js configuration file for the frontend.
func (s *PIrateRF) generateEnvJS() error {
	tmpl, err := template.New("envJS").Parse(envJSTemplate)
	if err != nil {
		return ctxerrors.Wrap(err, "failed to parse env.js template")
	}

	var buf bytes.Buffer

	templateData := map[string]string{
		"FilesDir": s.config.FilesDir,
	}

	if err := tmpl.Execute(&buf, templateData); err != nil {
		return ctxerrors.Wrap(err, "failed to execute env.js template")
	}

	envJSPath := path.Join(s.config.StaticDir, envJSFilename)
	if err := os.WriteFile(envJSPath, buf.Bytes(), filePerms); err != nil {
		return ctxerrors.Wrap(err, "failed to write env.js file")
	}

	logrus.Debugf("Generated env.js config file at %s", envJSPath)

	return nil
}
