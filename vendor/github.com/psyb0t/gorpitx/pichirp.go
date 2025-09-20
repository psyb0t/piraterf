package gorpitx

import (
	"encoding/json"
	"io"
	"strconv"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNamePICHIRP ModuleName = "pichirp"
)

type PICHIRP struct {
	// Frequency specifies the center frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// Bandwidth specifies the frequency sweep bandwidth in Hz. Required parameter.
	// Must be positive value.
	Bandwidth float64 `json:"bandwidth"`

	// Time specifies the sweep duration in seconds. Required parameter.
	// Must be positive value.
	Time float64 `json:"time"`
}

func (m *PICHIRP) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for pichirp binary.
func (m *PICHIRP) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args,
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add bandwidth argument (required)
	args = append(args,
		strconv.FormatFloat(m.Bandwidth, 'f', 0, 64))

	// Add time argument (required)
	args = append(args,
		strconv.FormatFloat(m.Time, 'f', -1, 64))

	return args
}

// validate validates all PICHIRP parameters.
func (m *PICHIRP) validate() error {
	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateBandwidth(); err != nil {
		return err
	}

	if err := m.validateTime(); err != nil {
		return err
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *PICHIRP) validateFrequency() error {
	if m.Frequency <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"frequency must be positive, got: %f",
			m.Frequency,
		)
	}

	// Validate frequency range using Hz-based validation
	if !isValidFreqHz(m.Frequency) {
		return ctxerrors.Wrapf(
			ErrFreqOutOfRange,
			"(%d kHz to %.0f MHz), got: %f Hz",
			minFreqKHz, getMaxFreqMHzDisplay(), m.Frequency,
		)
	}

	return nil
}

// validateBandwidth validates the bandwidth parameter.
func (m *PICHIRP) validateBandwidth() error {
	if m.Bandwidth <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"bandwidth must be positive, got: %f",
			m.Bandwidth,
		)
	}

	return nil
}

// validateTime validates the time parameter.
func (m *PICHIRP) validateTime() error {
	if m.Time <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"time must be positive, got: %f",
			m.Time,
		)
	}

	return nil
}
