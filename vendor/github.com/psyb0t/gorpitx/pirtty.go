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
	ModuleNamePIRTTY ModuleName = "pirtty"
)

const (
	defaultPIRTTYSpaceFrequency = 170
)

type PIRTTY struct {
	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// SpaceFrequency specifies the space frequency in Hz. Optional parameter.
	// Default: 170 Hz (mark frequency will be space + 170)
	SpaceFrequency *int `json:"spaceFrequency,omitempty"`

	// Message specifies the text message to transmit in RTTY. Required parameter.
	// Cannot be empty or whitespace only.
	Message string `json:"message"`
}

func (m *PIRTTY) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for pirtty
// binary.
func (m *PIRTTY) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args,
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add space frequency argument (default if not specified)
	spaceFreq := defaultPIRTTYSpaceFrequency
	if m.SpaceFrequency != nil {
		spaceFreq = *m.SpaceFrequency
	}

	args = append(args, strconv.Itoa(spaceFreq))

	// Add message argument (required)
	args = append(args, m.Message)

	return args
}

// validate validates all PIRTTY parameters.
func (m *PIRTTY) validate() error {
	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateSpaceFrequency(); err != nil {
		return err
	}

	if err := m.validateMessage(); err != nil {
		return err
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *PIRTTY) validateFrequency() error {
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

// validateSpaceFrequency validates the space frequency parameter.
func (m *PIRTTY) validateSpaceFrequency() error {
	if m.SpaceFrequency != nil && *m.SpaceFrequency <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"space frequency must be positive, got: %d",
			*m.SpaceFrequency,
		)
	}

	return nil
}

// validateMessage validates the message parameter.
func (m *PIRTTY) validateMessage() error {
	if strings.TrimSpace(m.Message) == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "message")
	}

	return nil
}
