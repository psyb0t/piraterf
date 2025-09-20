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
	ModuleNameSPECTRUMPAINT ModuleName = "spectrumpaint"
)

type SPECTRUMPAINT struct {
	// PictureFile specifies the path to the raw data file for spectrumpaint. Required parameter.
	// File must exist and be accessible. Should be raw data (320 bytes per row).
	PictureFile string `json:"pictureFile"`

	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// Excursion specifies the frequency excursion in Hz. Optional parameter.
	// Must be positive if specified. Default: 100000 Hz (100 kHz)
	Excursion *float64 `json:"excursion,omitempty"`
}

func (s *SPECTRUMPAINT) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, s); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := s.validate(); err != nil {
		return nil, nil, err
	}

	return s.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for spectrumpaint binary.
func (s *SPECTRUMPAINT) buildArgs() []string {
	var args []string

	// Add picture file argument (required)
	args = append(args, s.PictureFile)

	// Add frequency argument (required)
	args = append(args,
		strconv.FormatFloat(s.Frequency, 'f', 0, 64))

	// Add excursion argument (optional)
	if s.Excursion != nil {
		args = append(args,
			strconv.FormatFloat(*s.Excursion, 'f', 0, 64))
	}

	return args
}

// validate validates all SPECTRUMPAINT parameters.
func (s *SPECTRUMPAINT) validate() error {
	if err := s.validatePictureFile(); err != nil {
		return err
	}

	if err := s.validateFrequency(); err != nil {
		return err
	}

	if err := s.validateExcursion(); err != nil {
		return err
	}

	return nil
}

// validatePictureFile validates the picture file parameter.
func (s *SPECTRUMPAINT) validatePictureFile() error {
	if s.PictureFile == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "pictureFile")
	}

	if _, err := os.Stat(s.PictureFile); os.IsNotExist(err) {
		return ctxerrors.Wrapf(
			commonerrors.ErrFileNotFound,
			"file: %s",
			s.PictureFile,
		)
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (s *SPECTRUMPAINT) validateFrequency() error {
	if s.Frequency <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"frequency must be positive, got: %f",
			s.Frequency,
		)
	}

	// Validate frequency range using Hz-based validation
	if !isValidFreqHz(s.Frequency) {
		return ctxerrors.Wrapf(
			ErrFreqOutOfRange,
			"(%d kHz to %.0f MHz), got: %f Hz",
			minFreqKHz, getMaxFreqMHzDisplay(), s.Frequency,
		)
	}

	return nil
}

// validateExcursion validates the excursion parameter.
func (s *SPECTRUMPAINT) validateExcursion() error {
	if s.Excursion != nil && *s.Excursion <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"excursion must be positive, got: %f",
			*s.Excursion,
		)
	}

	return nil
}
