package gorpitx

import (
	"encoding/json"
	"io"
	"os"
	"slices"
	"strconv"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNameSENDIQ ModuleName = "sendiq"

	// IQ data type constants.
	IQTypeI16    = "i16"
	IQTypeU8     = "u8"
	IQTypeFloat  = "float"
	IQTypeDouble = "double"

	// Sample rate limits.
	// minSampleRate: Minimum supported sample rate.
	minSampleRate = 10000
	// maxSampleRate: Absolute maximum with decimation (10x MAX_SAMPLERATE).
	// Note: Native max is 200000 Hz; values above trigger automatic decimation.
	maxSampleRate = 2000000

	// Power level limits (clamped in sendiq.cpp lines 126-127).
	minPowerLevel = 0.0
	maxPowerLevel = 7.0

	// Default values from sendiq.cpp for optional parameters.
	// DefaultSampleRate: SampleRate=48000 (line 97).
	DefaultSampleRate = 48000
	// DefaultHarmonic: Harmonic=1 (line 100).
	DefaultHarmonic = 1
	// DefaultIQType: InputType=typeiq_i16 (line 102).
	DefaultIQType = IQTypeI16
	// DefaultPower: drivedds=0.1 (line 49).
	DefaultPower = 0.1
)

type SENDIQ struct {
	// InputFile specifies the input file path for I/Q samples. Required parameter.
	// Can be a file path or "-" for stdin (/dev/stdin).
	// File must exist before execution unless "-" is specified.
	InputFile string `json:"inputFile"`

	// Freq specifies the carrier frequency in Hz (NOT MHz). Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	// Example: 434000000 for 434 MHz
	Freq float64 `json:"freq"`

	// SampleRate specifies the sample rate in samples per second.
	// Optional parameter.
	// Range: 10,000 to 2,000,000 Hz
	// Default: 48000
	// Note: Rates > 200,000 will trigger automatic decimation
	SampleRate *int `json:"sampleRate,omitempty"`

	// Harmonic specifies the harmonic number. Optional parameter.
	// Must be positive integer (1, 2, 3, etc.)
	// Default: 1
	Harmonic *int `json:"harmonic,omitempty"`

	// IQType specifies the I/Q data type format. Optional parameter.
	// Valid values: "i16", "u8", "float", "double"
	// Default: "i16"
	// Note: When SharedMemToken is set, this is automatically forced to "float"
	IQType *string `json:"iqType,omitempty"`

	// Power specifies the power/drive level. Optional parameter.
	// Range: 0.0 to 7.0 (values outside range will be clamped)
	// Default: 0.1
	// Unit: Arbitrary drive level (not dBm or watts)
	Power *float64 `json:"power,omitempty"`

	// SharedMemToken specifies the shared memory token for IPC.
	// Optional parameter.
	// Must be non-zero integer
	// When set, automatically forces IQType to "float"
	// Enables runtime control via shared memory commands
	SharedMemToken *int `json:"sharedMemToken,omitempty"`

	// LoopMode enables continuous loop transmission. Optional parameter.
	// When true, seeks to beginning of file when EOF is reached
	// Default: false
	LoopMode bool `json:"loopMode,omitempty"`
}

func (m *SENDIQ) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for sendiq
// binary.
func (m *SENDIQ) buildArgs() []string {
	var args []string

	// Add input file argument (required)
	args = append(args, "-i", m.InputFile)

	// Add frequency argument (required, in Hz)
	args = append(args, "-f",
		strconv.FormatFloat(m.Freq, 'f', 0, 64))

	// Add sample rate argument (optional)
	if m.SampleRate != nil {
		args = append(args, "-s", strconv.Itoa(*m.SampleRate))
	}

	// Add harmonic argument (optional)
	if m.Harmonic != nil {
		args = append(args, "-h", strconv.Itoa(*m.Harmonic))
	}

	// Add IQ type argument (optional)
	// Note: If SharedMemToken is set, the binary forces float type anyway
	if m.IQType != nil {
		args = append(args, "-t", *m.IQType)
	}

	// Add power argument (optional)
	if m.Power != nil {
		args = append(args, "-p",
			strconv.FormatFloat(*m.Power, 'f', 2, 64))
	}

	// Add shared memory token argument (optional)
	if m.SharedMemToken != nil {
		args = append(args, "-m", strconv.Itoa(*m.SharedMemToken))
	}

	// Add loop mode flag
	if m.LoopMode {
		args = append(args, "-l")
	}

	return args
}

// validate validates all SENDIQ parameters.
func (m *SENDIQ) validate() error {
	if err := m.validateInputFile(); err != nil {
		return err
	}

	if err := m.validateFreq(); err != nil {
		return err
	}

	if err := m.validateSampleRate(); err != nil {
		return err
	}

	if err := m.validateHarmonic(); err != nil {
		return err
	}

	if err := m.validateIQType(); err != nil {
		return err
	}

	if err := m.validatePower(); err != nil {
		return err
	}

	if err := m.validateSharedMemToken(); err != nil {
		return err
	}

	return nil
}

// validateInputFile validates the input file parameter.
func (m *SENDIQ) validateInputFile() error {
	if m.InputFile == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "inputFile")
	}

	// Special case: "-" means stdin, which is always valid
	if m.InputFile == "-" {
		return nil
	}

	// Check if input file exists
	if _, err := os.Stat(m.InputFile); os.IsNotExist(err) {
		return ctxerrors.Wrapf(
			commonerrors.ErrFileNotFound,
			"input file: %s",
			m.InputFile,
		)
	}

	return nil
}

// validateFreq validates the frequency parameter.
func (m *SENDIQ) validateFreq() error {
	if m.Freq <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"frequency must be positive, got: %f",
			m.Freq,
		)
	}

	// Validate frequency range using Hz-based validation
	if !isValidFreqHz(m.Freq) {
		return ctxerrors.Wrapf(
			ErrFreqOutOfRange,
			"(%d kHz to %.0f MHz), got: %f Hz",
			minFreqKHz, getMaxFreqMHzDisplay(), m.Freq,
		)
	}

	return nil
}

// validateSampleRate validates the sample rate parameter.
func (m *SENDIQ) validateSampleRate() error {
	if m.SampleRate == nil {
		return nil
	}

	if *m.SampleRate < minSampleRate {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"sample rate must be at least %d, got: %d",
			minSampleRate, *m.SampleRate,
		)
	}

	if *m.SampleRate > maxSampleRate {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"sample rate must be at most %d, got: %d",
			maxSampleRate, *m.SampleRate,
		)
	}

	return nil
}

// validateHarmonic validates the harmonic parameter.
func (m *SENDIQ) validateHarmonic() error {
	if m.Harmonic == nil {
		return nil
	}

	if *m.Harmonic < 1 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"harmonic must be positive, got: %d",
			*m.Harmonic,
		)
	}

	return nil
}

// validateIQType validates the IQ type parameter.
func (m *SENDIQ) validateIQType() error {
	if m.IQType == nil {
		return nil
	}

	// Check against valid IQ types
	validTypes := []string{IQTypeI16, IQTypeU8, IQTypeFloat, IQTypeDouble}
	if !slices.Contains(validTypes, *m.IQType) {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"IQ type must be one of [i16, u8, float, double], got: %s",
			*m.IQType,
		)
	}

	return nil
}

// validatePower validates the power parameter.
func (m *SENDIQ) validatePower() error {
	if m.Power == nil {
		return nil
	}

	// Clamp power to valid range
	if *m.Power < minPowerLevel {
		*m.Power = minPowerLevel
	}

	if *m.Power > maxPowerLevel {
		*m.Power = maxPowerLevel
	}

	return nil
}

// validateSharedMemToken validates the shared memory token parameter.
func (m *SENDIQ) validateSharedMemToken() error {
	if m.SharedMemToken == nil {
		return nil
	}

	if *m.SharedMemToken == 0 {
		return ctxerrors.Wrap(
			commonerrors.ErrInvalidValue,
			"shared memory token must be non-zero",
		)
	}

	return nil
}
