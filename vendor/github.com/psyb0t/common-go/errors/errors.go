package commonerrors

import "errors"

var (
	// Configuration & Environment errors
	ErrEnvVarNotSet              = errors.New("env var is not set")
	ErrRequiredConfigValueNotSet = errors.New("required config value is not set")
	ErrEmptyMigrationsPath       = errors.New("migrations path is empty")

	// File & Path errors
	ErrFileInvalid           = errors.New("invalid file")
	ErrFileNotFound          = errors.New("file not found")
	ErrPathIsRequired        = errors.New("path is required")
	ErrCouldNotDownloadFiles = errors.New("could not download files")

	// Validation & Input errors
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrInvalidValue     = errors.New("invalid value")
	ErrTargetNotPointer = errors.New("target is not a pointer")
	ErrCouldNotDecode   = errors.New("could not decode")

	// Field & Data errors
	ErrNilOutput                      = errors.New("output is nil")
	ErrNilRequestBody                 = errors.New("request body is nil")
	ErrNilField                       = errors.New("field is nil")
	ErrRequiredFieldNotSet            = errors.New("required field is not set")
	ErrRequiredLLMResponseFieldNotSet = errors.New("required llm response field is not set")

	// Job & Process errors
	ErrJobFailed                 = errors.New("job failed")
	ErrUnexpectedNumberOfResults = errors.New("unexpected number of results")
	ErrNotFound                  = errors.New("not found")

	// Process State errors
	ErrFailed     = errors.New("failed")
	ErrTimeout    = errors.New("timeout")
	ErrTerminated = errors.New("terminated")
	ErrKilled     = errors.New("killed")
	ErrClosing    = errors.New("closing")

	// API & HTTP errors
	ErrAPIError                 = errors.New("API error")
	ErrAPIKeyNotSet             = errors.New("api key is not set")
	ErrUnexpectedHTTPStatusCode = errors.New("unexpected http status code")
	ErrNotAuthenticated         = errors.New("not authenticated")

	// TLS & Security errors
	ErrTLSCertFileNotSpecified = errors.New("TLS cert file not specified")
	ErrTLSKeyFileNotSpecified  = errors.New("TLS key file not specified")
)
