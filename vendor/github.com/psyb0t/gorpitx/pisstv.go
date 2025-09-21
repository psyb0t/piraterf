package gorpitx

import (
	"encoding/json"
	"io"
	"os"
	"strconv"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNamePISSSTV ModuleName = "pisstv"
)

type PISSTV struct {
	// PictureFile specifies the .rgb picture file to transmit. Required parameter.
	// File must be exactly 320 pixels wide, any height, RGB format (3 bytes per pixel).
	PictureFile string `json:"pictureFile"`

	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`
}

func (m *PISSTV) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for pisstv binary.
func (m *PISSTV) buildArgs() []string {
	var args []string

	// Add picture file argument (required)
	args = append(args, m.PictureFile)

	// Add frequency argument (required)
	args = append(args, strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	return args
}

// validate validates all PISSTV parameters.
func (m *PISSTV) validate() error {
	if err := m.validatePictureFile(); err != nil {
		return err
	}

	if err := m.validateFrequency(); err != nil {
		return err
	}

	return nil
}

// validatePictureFile validates the picture file parameter.
func (m *PISSTV) validatePictureFile() error {
	if m.PictureFile == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "pictureFile")
	}

	// Check if picture file exists
	if _, err := os.Stat(m.PictureFile); os.IsNotExist(err) {
		return ctxerrors.Wrapf(
			commonerrors.ErrFileNotFound,
			"picture file: %s",
			m.PictureFile,
		)
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *PISSTV) validateFrequency() error {
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
