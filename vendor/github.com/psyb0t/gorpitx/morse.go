package gorpitx

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNameMORSE ModuleName = "morse"
)

type MORSE struct {
	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// Rate specifies the transmission rate in dits per minute. Required parameter.
	// Must be positive integer value.
	Rate int `json:"rate"`

	// Message specifies the text message to transmit in Morse code. Required
	// parameter.
	// Cannot be empty or whitespace only.
	Message string `json:"message"`
}

func (m *MORSE) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for morse
// binary.
func (m *MORSE) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args,
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add rate argument (required)
	args = append(args, strconv.Itoa(m.Rate))

	// Add message argument (required)
	args = append(args, m.Message)

	return args
}

// validate validates all MORSE parameters.
func (m *MORSE) validate() error {
	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateRate(); err != nil {
		return err
	}

	if err := m.validateMessage(); err != nil {
		return err
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *MORSE) validateFrequency() error {
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

// validateRate validates the rate parameter.
func (m *MORSE) validateRate() error {
	if m.Rate <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"morse rate must be positive, got: %d",
			m.Rate,
		)
	}

	return nil
}

// validateMessage validates the message parameter.
func (m *MORSE) validateMessage() error {
	if strings.TrimSpace(m.Message) == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "message")
	}

	return nil
}
