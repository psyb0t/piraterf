package gorpitx

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNameFSK ModuleName = "fsk"
)

const (
	defaultFSKBaudRate = 50
)

// InputType defines the type of input for FSK transmission.
type InputType = string

const (
	InputTypeFile InputType = "file"
	InputTypeText InputType = "text"
)

type FSK struct {
	// InputType specifies whether input is from file or text. Required parameter.
	// Must be either "file" or "text".
	InputType InputType `json:"inputType"`

	// File specifies the path to input file. Required when InputType is "file".
	// Cannot be specified when InputType is "text".
	File string `json:"file,omitempty"`

	// Text specifies the input text to transmit. Required when InputType is
	// "text". Cannot be specified when InputType is "file".
	Text string `json:"text,omitempty"`

	// BaudRate specifies the transmission baud rate. Optional parameter.
	// Default: 50 baud (cleanest in testing with rpitx FSK transmission)
	BaudRate *int `json:"baudRate,omitempty"`

	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`
}

func (m *FSK) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	stdin, err := m.prepareStdin()
	if err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), stdin, nil
}

// buildArgs converts the struct fields into command-line arguments for FSK
// script.
func (m *FSK) buildArgs() []string {
	var args []string

	// Add baud rate argument (default if not specified)
	baudRate := defaultFSKBaudRate
	if m.BaudRate != nil {
		baudRate = *m.BaudRate
	}

	args = append(args, strconv.Itoa(baudRate))

	// Add frequency argument (required)
	args = append(args, strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	return args
}

// prepareStdin prepares the stdin reader based on input type.
func (m *FSK) prepareStdin() (io.Reader, error) {
	var baseReader io.Reader

	switch m.InputType {
	case InputTypeText:
		baseReader = strings.NewReader(m.Text)
	case InputTypeFile:
		file, err := os.Open(m.File)
		if err != nil {
			return nil, ctxerrors.Wrapf(
				err,
				"failed to open file: %s",
				m.File,
			)
		}

		baseReader = file
	default:
		return nil, ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"invalid input type: %s",
			m.InputType,
		)
	}

	return io.MultiReader(
		baseReader,
		strings.NewReader("\n"),
	), nil
}

// validate validates all FSK parameters.
func (m *FSK) validate() error {
	if err := m.validateInputType(); err != nil {
		return err
	}

	if err := m.validateInputFields(); err != nil {
		return err
	}

	if err := m.validateBaudRate(); err != nil {
		return err
	}

	if err := m.validateFrequency(); err != nil {
		return err
	}

	return nil
}

// validateInputType validates the input type parameter.
func (m *FSK) validateInputType() error {
	if m.InputType == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "inputType")
	}

	if m.InputType != InputTypeFile && m.InputType != InputTypeText {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"inputType must be 'file' or 'text', got: %s",
			m.InputType,
		)
	}

	return nil
}

// validateInputFields validates file/text fields based on input type.
func (m *FSK) validateInputFields() error {
	switch m.InputType {
	case InputTypeFile:
		if strings.TrimSpace(m.File) == "" {
			return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "file")
		}

		// Check if file exists
		if _, err := os.Stat(m.File); os.IsNotExist(err) {
			return ctxerrors.Wrapf(
				commonerrors.ErrFileNotFound,
				"input file: %s",
				m.File,
			)
		}
	case InputTypeText:
		if strings.TrimSpace(m.Text) == "" {
			return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "text")
		}
	}

	return nil
}

// validateBaudRate validates the baud rate parameter.
func (m *FSK) validateBaudRate() error {
	if m.BaudRate != nil && *m.BaudRate <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"baud rate must be positive, got: %d",
			*m.BaudRate,
		)
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *FSK) validateFrequency() error {
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
