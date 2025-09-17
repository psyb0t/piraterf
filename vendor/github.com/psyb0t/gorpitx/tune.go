package gorpitx

import (
	"encoding/json"
	"strconv"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNameTUNE ModuleName = "tune"
)

type TUNE struct {
	// `-f` specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency *float64 `json:"frequency,omitempty"`

	// `-e` flag exits immediately without killing the carrier.
	// Optional parameter, defaults to false.
	ExitImmediate *bool `json:"exitImmediate,omitempty"`

	// `-p` specifies clock PPM correction instead of NTP adjust.
	// Optional parameter, must be positive if provided.
	PPM *float64 `json:"ppm,omitempty"`
}

func (m *TUNE) ParseArgs(args json.RawMessage) ([]string, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, err
	}

	return m.buildArgs(), nil
}

// buildArgs converts the struct fields into command-line arguments for tune binary.
func (m *TUNE) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args, "-f",
		strconv.FormatFloat(*m.Frequency, 'f', 0, 64))

	// Add exit immediate flag
	if m.ExitImmediate != nil && *m.ExitImmediate {
		args = append(args, "-e")
	}

	// Add PPM argument
	if m.PPM != nil {
		args = append(args, "-p",
			strconv.FormatFloat(*m.PPM, 'f', -1, 64))
	}

	return args
}

// validate validates all TUNE parameters.
func (m *TUNE) validate() error {
	if err := m.validateFreq(); err != nil {
		return err
	}

	if err := m.validatePPM(); err != nil {
		return err
	}

	return nil
}

// validateFreq validates the frequency parameter.
func (m *TUNE) validateFreq() error {
	// Frequency is required
	if m.Frequency == nil {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "frequency")
	}

	if *m.Frequency <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"frequency must be positive, got: %f",
			*m.Frequency,
		)
	}

	// Validate frequency range using Hz-based validation
	if !isValidFreqHz(*m.Frequency) {
		return ctxerrors.Wrapf(
			ErrFreqOutOfRange,
			"(%d kHz to %.0f MHz), got: %f Hz",
			minFreqKHz, getMaxFreqMHzDisplay(), *m.Frequency,
		)
	}

	return nil
}

// validatePPM validates the PPM parameter.
func (m *TUNE) validatePPM() error {
	// PPM is optional, but if provided must be positive
	if m.PPM != nil && *m.PPM <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"PPM must be positive, got: %f",
			*m.PPM,
		)
	}

	return nil
}
