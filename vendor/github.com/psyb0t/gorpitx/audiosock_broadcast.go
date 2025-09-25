package gorpitx

import (
	"encoding/json"
	"io"
	"slices"
	"strconv"

	commonerrors "github.com/psyb0t/common-go/errors"
	"github.com/psyb0t/ctxerrors"
)

const (
	ModuleNameAudioSockBroadcast ModuleName = "audiosock-broadcast"
)

type ModulationType = string

const (
	ModulationAM  ModulationType = "AM"
	ModulationDSB ModulationType = "DSB"
	ModulationUSB ModulationType = "USB"
	ModulationLSB ModulationType = "LSB"
	ModulationFM  ModulationType = "FM"
	ModulationRAW ModulationType = "RAW"
)

const (
	defaultAudioSockBroadcastSampleRate = 48000
)

type AudioSockBroadcast struct {
	// SocketPath specifies the Unix socket path for audio input. Required.
	SocketPath string `json:"socketPath"`

	// Frequency specifies the carrier frequency in Hz. Required parameter.
	// Range: 50 kHz to 1500 MHz (50000 to 1500000000 Hz)
	Frequency float64 `json:"frequency"`

	// SampleRate specifies the audio sample rate. Optional parameter.
	// Default: 48000 Hz
	SampleRate *int `json:"sampleRate,omitempty"`

	// Modulation specifies the modulation type. Optional parameter.
	// If not specified, uses default "FM".
	// Available: AM, DSB, USB, LSB, FM, RAW
	Modulation *string `json:"modulation,omitempty"`

	// Gain specifies the gain multiplier for the audio signal. Optional parameter.
	// Default: 1.0
	Gain *float64 `json:"gain,omitempty"`
}

func (m *AudioSockBroadcast) ParseArgs(
	args json.RawMessage,
) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(err, "failed to unmarshal args")
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for
// AudioSock script.
func (m *AudioSockBroadcast) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args,
		strconv.FormatFloat(m.Frequency, 'f', 0, 64))

	// Add socket path argument (required)
	args = append(args, m.SocketPath)

	// Add sample rate argument (default if not specified)
	sampleRate := defaultAudioSockBroadcastSampleRate
	if m.SampleRate != nil {
		sampleRate = *m.SampleRate
	}

	args = append(args, strconv.Itoa(sampleRate))

	// Add modulation argument (default if not specified)
	modulation := ModulationFM
	if m.Modulation != nil {
		modulation = *m.Modulation
	}

	args = append(args, modulation)

	// Add gain argument (default if not specified)
	gain := 1.0
	if m.Gain != nil {
		gain = *m.Gain
	}

	args = append(args, strconv.FormatFloat(gain, 'f', -1, 64))

	return args
}

// validate validates all AudioSock parameters.
func (m *AudioSockBroadcast) validate() error {
	if err := m.validateSocketPath(); err != nil {
		return err
	}

	if err := m.validateFrequency(); err != nil {
		return err
	}

	if err := m.validateSampleRate(); err != nil {
		return err
	}

	if err := m.validateModulation(); err != nil {
		return err
	}

	if err := m.validateGain(); err != nil {
		return err
	}

	return nil
}

// validateSocketPath validates the socket path parameter.
func (m *AudioSockBroadcast) validateSocketPath() error {
	if m.SocketPath == "" {
		return ctxerrors.Wrap(
			commonerrors.ErrRequiredFieldNotSet, "socketPath")
	}

	return nil
}

// validateFrequency validates the frequency parameter.
func (m *AudioSockBroadcast) validateFrequency() error {
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

// validateSampleRate validates the sample rate parameter.
func (m *AudioSockBroadcast) validateSampleRate() error {
	if m.SampleRate != nil && *m.SampleRate <= 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"sample rate must be positive, got: %d",
			*m.SampleRate,
		)
	}

	return nil
}

// validateModulation validates the modulation parameter.
func (m *AudioSockBroadcast) validateModulation() error {
	if m.Modulation == nil {
		return nil // Optional parameter
	}

	validModulations := []ModulationType{
		ModulationAM,
		ModulationDSB,
		ModulationUSB,
		ModulationLSB,
		ModulationFM,
		ModulationRAW,
	}

	modulation := *m.Modulation
	if slices.Contains(validModulations, modulation) {
		return nil
	}

	return ctxerrors.Wrapf(
		commonerrors.ErrInvalidValue,
		"invalid modulation: %s, valid modulations: %v",
		modulation, validModulations,
	)
}

// validateGain validates the gain parameter.
func (m *AudioSockBroadcast) validateGain() error {
	if m.Gain != nil && *m.Gain < 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"gain must be non-negative, got: %f",
			*m.Gain,
		)
	}

	return nil
}
