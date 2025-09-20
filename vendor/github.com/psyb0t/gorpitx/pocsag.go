package gorpitx

import (
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNamePOCSAG ModuleName = "pocsag"
)

type POCSAG struct {
	// `-f` specifies the frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// `-r` specifies the baud rate. Optional, must be 512, 1200, or 2400.
	// Defaults to 1200 baud.
	BaudRate *int `json:"baudRate,omitempty"`

	// `-b` specifies the function bits. Optional, must be 0-3.
	// Defaults to 3.
	FunctionBits *int `json:"functionBits,omitempty"`

	// `-n` flag enables numeric mode. Optional, defaults to false.
	NumericMode *bool `json:"numericMode,omitempty"`

	// `-t` specifies the repeat count. Optional, defaults to 4.
	RepeatCount *int `json:"repeatCount,omitempty"`

	// `-i` flag inverts polarity. Optional, defaults to false.
	InvertPolarity *bool `json:"invertPolarity,omitempty"`

	// `-d` flag enables debug mode. Optional, defaults to false.
	Debug *bool `json:"debug,omitempty"`

	// Messages array specifies the address:message pairs to transmit.
	// Required, must have at least one message.
	Messages []POCSAGMessage `json:"messages"`
}

type POCSAGMessage struct {
	// Address specifies the pager address. Required.
	Address int `json:"address"`

	// Message specifies the message text to transmit. Required.
	Message string `json:"message"`

	// FunctionBits optionally overrides the global function bits for this message.
	FunctionBits *int `json:"functionBits,omitempty"`
}

func (m *POCSAG) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	cmdArgs := m.buildArgs()
	stdin := m.buildStdin()

	return cmdArgs, stdin, nil
}

// buildArgs converts the struct fields into command-line arguments for pocsag binary.
func (m *POCSAG) buildArgs() []string {
	args := make([]string, 0)

	// Add frequency argument
	args = append(args, "-f",
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add baud rate argument
	if m.BaudRate != nil {
		args = append(args, "-r",
			strconv.Itoa(*m.BaudRate))
	}

	// Add function bits argument
	if m.FunctionBits != nil {
		args = append(args, "-b",
			strconv.Itoa(*m.FunctionBits))
	}

	// Add numeric mode flag
	if m.NumericMode != nil && *m.NumericMode {
		args = append(args, "-n")
	}

	// Add repeat count argument
	if m.RepeatCount != nil {
		args = append(args, "-t",
			strconv.Itoa(*m.RepeatCount))
	}

	// Add invert polarity flag
	if m.InvertPolarity != nil && *m.InvertPolarity {
		args = append(args, "-i")
	}

	// Add debug flag
	if m.Debug != nil && *m.Debug {
		args = append(args, "-d")
	}

	return args
}

// buildStdin converts messages to stdin format expected by pocsag binary.
func (m *POCSAG) buildStdin() io.Reader {
	lines := make([]string, 0, len(m.Messages))

	for _, msg := range m.Messages {
		// Format: address:message
		msgStr := strconv.Itoa(msg.Address) + ":" + msg.Message
		lines = append(lines, msgStr)
	}

	// Join with newlines and create a string reader
	stdinContent := strings.Join(lines, "\n")

	return strings.NewReader(stdinContent)
}

// validate validates all POCSAG parameters.
func (m *POCSAG) validate() error {
	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateBaudRate(); err != nil {
		return err
	}

	if err := m.validateFunctionBits(); err != nil {
		return err
	}

	if err := m.validateRepeatCount(); err != nil {
		return err
	}

	if err := m.validateMessages(); err != nil {
		return err
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *POCSAG) validateFrequency() error {
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

// validateBaudRate validates the baud rate parameter.
func (m *POCSAG) validateBaudRate() error {
	// Baud rate is optional
	if m.BaudRate == nil {
		return nil
	}

	// Must be one of the valid baud rates
	validRates := []int{512, 1200, 2400}
	if slices.Contains(validRates, *m.BaudRate) {
		return nil
	}

	return ctxerrors.Wrapf(
		commonerrors.ErrInvalidValue,
		"baud rate must be 512, 1200, or 2400, got: %d",
		*m.BaudRate,
	)
}

// validateFunctionBits validates the function bits parameter.
func (m *POCSAG) validateFunctionBits() error {
	// Function bits is optional
	if m.FunctionBits == nil {
		return nil
	}

	if *m.FunctionBits < 0 || *m.FunctionBits > 3 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"function bits must be 0-3, got: %d",
			*m.FunctionBits,
		)
	}

	return nil
}

// validateRepeatCount validates the repeat count parameter.
func (m *POCSAG) validateRepeatCount() error {
	// Repeat count is optional
	if m.RepeatCount == nil {
		return nil
	}

	if *m.RepeatCount <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"repeat count must be positive, got: %d",
			*m.RepeatCount,
		)
	}

	return nil
}

// validateMessages validates the messages array.
func (m *POCSAG) validateMessages() error {
	// Messages array is required
	if len(m.Messages) == 0 {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "messages")
	}

	// Validate each message
	for i, msg := range m.Messages {
		if err := m.validateMessage(msg, i); err != nil {
			return err
		}
	}

	return nil
}

// validateMessage validates a single POCSAG message.
func (m *POCSAG) validateMessage(msg POCSAGMessage, index int) error {
	// Address must be non-negative
	if msg.Address < 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"message[%d].address must be non-negative, got: %d",
			index, msg.Address,
		)
	}

	// Message text cannot be empty
	if strings.TrimSpace(msg.Message) == "" {
		return ctxerrors.Wrapf(
			commonerrors.ErrRequiredFieldNotSet,
			"message[%d].message",
			index,
		)
	}

	// Validate per-message function bits if specified
	if msg.FunctionBits != nil {
		if *msg.FunctionBits < 0 || *msg.FunctionBits > 3 {
			return ctxerrors.Wrapf(
				commonerrors.ErrInvalidValue,
				"message[%d].functionBits must be 0-3, got: %d",
				index, *msg.FunctionBits,
			)
		}
	}

	return nil
}
