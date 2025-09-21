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
	ModuleNamePIFMRDS ModuleName = "pifmrds"

	piCodeLength = 4  // PI code must be 4 hex digits
	psMaxLength  = 8  // PS text maximum 8 characters
	rtMaxLength  = 64 // RT text maximum 64 characters
)

type PIFMRDS struct {
	// `-freq` specifies the carrier frequency (in MHz). Example: `-freq 107.9`.
	// This is what frequency people tune to on their radios.
	Freq float64 `json:"freq,omitempty"`

	// `-audio` specifies an audio file to play as audio. The sample rate does
	// not matter: Pi-FM-RDS will resample and filter it. If a stereo file is
	// provided, Pi-FM-RDS will produce an FM-Stereo signal. Example:
	// `-audio sound.wav`. The supported formats depend on `libsndfile`. This
	// includes WAV and Ogg/Vorbis (among others) but not MP3. Specify `-` as
	// the file name to read audio data on standard input.
	Audio string `json:"audio,omitempty"`

	// `-pi` specifies the PI-code of the RDS broadcast. 4 hexadecimal digits.
	// Example: `-pi FFFF`. This is the internal station ID that RDS radios use
	// to identify your station.
	PI string `json:"pi,omitempty"`

	// `-ps` specifies the station name (Program Service name, PS) of the RDS
	// broadcast. Limit: 8 characters. Example: `-ps RASP-PI`. This is the
	// STATION NAME that appears on car radios and RDS displays. By default the
	// PS changes back and forth between `Pi-FmRds` and a sequence number,
	// starting at `00000000`. The PS changes around one time per second.
	PS string `json:"ps,omitempty"`

	// `-rt` specifies the radiotext (RT) to be transmitted. Limit: 64
	// characters. Example: `-rt 'Hello, world!'`. This is the scrolling text
	// message shown on RDS displays.
	RT string `json:"rt,omitempty"`

	// `-ppm` specifies your Raspberry Pi's oscillator error in parts per
	// million (ppm).
	// Compensates for Raspberry Pi clock inaccuracy (usually 0 is fine).
	PPM *float64 `json:"ppm,omitempty"`

	// `-ctl` specifies a named pipe (FIFO) to use as a control channel to
	// change PS and RT at run-time. Create with "mkfifo /tmp/rds_ctl" then
	// echo commands like "PS New Name".
	ControlPipe *string `json:"controlPipe,omitempty"`
}

func (m *PIFMRDS) ParseArgs(args json.RawMessage) ([]string, io.Reader, error) {
	if err := json.Unmarshal(args, m); err != nil {
		return nil, nil, ctxerrors.Wrap(
			err,
			"failed to unmarshal args",
		)
	}

	if err := m.validate(); err != nil {
		return nil, nil, err
	}

	return m.buildArgs(), nil, nil
}

// buildArgs converts the struct fields into command-line arguments for
// pifmrds binary.
func (m *PIFMRDS) buildArgs() []string {
	var args []string

	// Add frequency argument (required)
	args = append(args, "-freq",
		strconv.FormatFloat(m.Freq, 'f', 1, 64))

	// Add audio argument (required)
	args = append(args, "-audio", m.Audio)

	// Add PI argument
	if m.PI != "" {
		args = append(args, "-pi", m.PI)
	}

	// Add PS argument
	if m.PS != "" {
		args = append(args, "-ps", m.PS)
	}

	// Add RT argument
	if m.RT != "" {
		args = append(args, "-rt", m.RT)
	}

	// Add PPM argument
	if m.PPM != nil {
		args = append(args, "-ppm",
			strconv.FormatFloat(*m.PPM, 'f', -1, 64))
	}

	// Add control pipe argument
	if m.ControlPipe != nil && *m.ControlPipe != "" {
		args = append(args, "-ctl", *m.ControlPipe)
	}

	return args
}

// validate validates all PIFMRDSArgs parameters.
func (m *PIFMRDS) validate() error {
	if err := m.validateFreq(); err != nil {
		return err
	}

	if err := m.validateAudio(); err != nil {
		return err
	}

	if err := m.validatePI(); err != nil {
		return err
	}

	if err := m.validatePS(); err != nil {
		return err
	}

	if err := m.validateRT(); err != nil {
		return err
	}

	if err := m.validatePPM(); err != nil {
		return err
	}

	if err := m.validateControlPipe(); err != nil {
		return err
	}

	return nil
}

// validateFreq validates the frequency parameter.
func (m *PIFMRDS) validateFreq() error {
	// Validate required frequency
	if m.Freq == 0 {
		return ctxerrors.Wrap(
			commonerrors.ErrRequiredFieldNotSet,
			"freq",
		)
	}

	if m.Freq < 0 {
		return ctxerrors.Wrapf(
			commonerrors.ErrInvalidValue,
			"frequency must be positive, got: %f",
			m.Freq,
		)
	}

	// RPiTX frequency range validation using utility functions
	// Convert MHz to Hz for validation since isValidFreqHz expects Hz
	freqHz := mHzToHz(m.Freq)
	if !isValidFreqHz(freqHz) {
		return ctxerrors.Wrapf(
			ErrFreqOutOfRange,
			"(%d kHz to %.0f MHz), got: %f",
			minFreqKHz, getMaxFreqMHzDisplay(), m.Freq,
		)
	}

	// Validate frequency precision (pifmrds works best with 1 decimal place)
	if !hasValidFreqPrecision(m.Freq) {
		return ctxerrors.Wrapf(
			ErrFreqPrecision,
			"(0.1 MHz precision), got: %f",
			m.Freq,
		)
	}

	return nil
}

// validateAudio validates the audio parameter.
func (m *PIFMRDS) validateAudio() error {
	// Audio file is required
	if m.Audio == "" {
		return ctxerrors.Wrap(commonerrors.ErrRequiredFieldNotSet, "audio")
	}

	// Check if audio file exists (no stdin support for now)
	if _, err := os.Stat(m.Audio); os.IsNotExist(err) {
		return ctxerrors.Wrapf(
			commonerrors.ErrFileNotFound,
			"file: %s",
			m.Audio,
		)
	}

	return nil
}

// validatePI validates the PI code parameter.
func (m *PIFMRDS) validatePI() error {
	// Validate PI code (4 hex digits) if not empty
	if m.PI != "" {
		pi := strings.TrimSpace(m.PI)
		if len(pi) != piCodeLength {
			return ctxerrors.Wrapf(
				commonerrors.ErrInvalidValue,
				"PI code must be exactly 4 characters, got: %s",
				pi,
			)
		}

		if _, err := strconv.ParseUint(pi, 16, 16); err != nil {
			return ctxerrors.Wrapf(
				ErrPIInvalidHex, "got: %s", pi)
		}
	}

	return nil
}

// validatePS validates the Program Service name parameter.
func (m *PIFMRDS) validatePS() error {
	// Validate PS (Program Service name - 8 chars max) if not empty
	if m.PS != "" {
		if len(m.PS) > psMaxLength {
			return ctxerrors.Wrapf(
				ErrPSTooLong,
				"got: %d chars",
				len(m.PS),
			)
		}

		if strings.TrimSpace(m.PS) == "" {
			return ctxerrors.Wrap(
				commonerrors.ErrInvalidValue,
				"PS text cannot be empty when specified",
			)
		}
	}

	return nil
}

// validateRT validates the Radio Text parameter.
func (m *PIFMRDS) validateRT() error {
	// Validate RT (Radio Text - 64 chars max) if not empty
	if m.RT != "" {
		if len(m.RT) > rtMaxLength {
			return ctxerrors.Wrapf(
				commonerrors.ErrInvalidValue,
				"RT text must be 64 characters or less, got: %d chars",
				len(m.RT),
			)
		}
	}

	return nil
}

// validatePPM validates the PPM parameter.
func (m *PIFMRDS) validatePPM() error {
	// PPM can be any float value (positive, negative, or zero)
	// No validation needed for PPM
	return nil
}

// validateControlPipe validates the control pipe parameter.
func (m *PIFMRDS) validateControlPipe() error {
	// Validate optional control pipe path
	if m.ControlPipe != nil {
		pipe := strings.TrimSpace(*m.ControlPipe)
		if pipe == "" {
			return ctxerrors.Wrap(
				commonerrors.ErrInvalidValue,
				"control pipe path cannot be empty when specified",
			)
		}

		// Check if the control pipe exists (must be created with mkfifo first)
		if _, err := os.Stat(pipe); os.IsNotExist(err) {
			return ctxerrors.Wrapf(
				commonerrors.ErrFileNotFound,
				"control pipe does not exist: %s (create with: mkfifo %s)",
				pipe, pipe,
			)
		}
	}

	return nil
}
