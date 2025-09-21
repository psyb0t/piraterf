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
	ModuleNameFT8 ModuleName = "pift8"

	ft8OffsetMin     = 0    // Minimum frequency offset in Hz
	ft8OffsetMax     = 2500 // Maximum frequency offset in Hz
	ft8OffsetDefault = 1240 // Default frequency offset in Hz
)

type FT8 struct {
	// `-f` specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// `-m` specifies the message to transmit. Required parameter.
	// Example: "CQ CA0ALL JN06"
	Message string `json:"message"`

	// `-p` specifies clock PPM correction instead of NTP adjust.
	// Optional parameter, defaults to automatic NTP adjustment.
	PPM *float64 `json:"ppm,omitempty"`

	// `-o` specifies frequency offset (0-2500Hz). Optional parameter.
	// Default: 1240Hz
	Offset *float64 `json:"offset,omitempty"`

	// `-s` specifies time slot to transmit (0 or 1). Optional parameter.
	// 0 = first 15s slot, 1 = second 15s slot, 2 = always (every 15s)
	// Default: 0
	Slot *int `json:"slot,omitempty"`

	// `-r` flag enables repeat mode (every 15s). Optional parameter.
	// Default: false (single transmission)
	Repeat *bool `json:"repeat,omitempty"`
}

func (m *FT8) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for pift8
// binary.
func (m *FT8) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args, "-f",
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add message argument (required)
	args = append(args, "-m", m.Message)

	// Add PPM argument
	if m.PPM != nil {
		args = append(args, "-p",
			strconv.FormatFloat(*m.PPM, 'f', -1, 64))
	}

	// Add offset argument
	if m.Offset != nil {
		args = append(args, "-o",
			strconv.FormatFloat(*m.Offset, 'f', 0, 64))
	}

	// Add slot argument
	if m.Slot != nil {
		args = append(args, "-s", strconv.Itoa(*m.Slot))
	}

	// Add repeat flag
	if m.Repeat != nil && *m.Repeat {
		args = append(args, "-r")
	}

	return args
}

// validate validates all FT8 parameters.
func (m *FT8) validate() error {
	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateMessage(); err != nil {
		return err
	}

	if err := m.validatePPM(); err != nil {
		return err
	}

	if err := m.validateOffset(); err != nil {
		return err
	}

	if err := m.validateSlot(); err != nil {
		return err
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *FT8) validateFrequency() error {
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

// validateMessage validates the message parameter.
func (m *FT8) validateMessage() error {
	if strings.TrimSpace(m.Message) == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "message")
	}

	return nil
}

// validatePPM validates the PPM parameter.
func (m *FT8) validatePPM() error {
	// PPM can be any float value (positive, negative, or zero)
	// No validation needed for PPM
	return nil
}

// validateOffset validates the offset parameter.
func (m *FT8) validateOffset() error {
	if m.Offset != nil {
		if *m.Offset < ft8OffsetMin || *m.Offset > ft8OffsetMax {
			return ctxerrors.Wrapf(
				commonerrors.ErrInvalidValue,
				"FT8 offset must be between %d and %d Hz, got: %f",
				ft8OffsetMin, ft8OffsetMax, *m.Offset,
			)
		}
	}

	return nil
}

// validateSlot validates the slot parameter.
func (m *FT8) validateSlot() error {
	if m.Slot != nil {
		if *m.Slot < 0 || *m.Slot > 2 {
			return ctxerrors.Wrapf(
				commonerrors.ErrInvalidValue,
				"FT8 slot must be 0, 1, or 2, got: %d",
				*m.Slot,
			)
		}
	}

	return nil
}
