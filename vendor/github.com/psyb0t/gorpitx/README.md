# gorpitx

🚀 **Go wrapper that fucking executes rpitx modules without the bullshit.**

Tired of wrestling with raw C binaries like a goddamn caveman? This badass Go interface wraps rpitx so you can transmit radio signals without losing your shit. Singleton pattern because we're not animals, and proper process management because segfaults are for scrubs.

## 📡 What This Bastard Does

Executes rpitx modules through Go without the usual clusterfuck of manual process wrangling. Supports dev mode (fake transmission for testing) and production mode (actual RF carnage).

**Modules:**

- **pifmrds**: FM broadcasting with RDS data (frequency in MHz)
- **tune**: Simple carrier wave generation (frequency in Hz)
- **pichirp**: Carrier wave sweep generator (frequency in Hz)
- **morse**: Morse code transmission (frequency in Hz)
- **pocsag**: Pager protocol transmission (frequency in Hz)
- **spectrumpaint**: Spectrum painting transmission (frequency in Hz)
- **pift8**: FT8 digital mode transmission (frequency in Hz)

**Architecture Highlights:**

- Singleton pattern with `GetInstance()` because global state done right
- Module interface for adding more transmission types without breaking shit
- Process management with timeout and graceful stop (no zombie apocalypse)
- Dev mode with mock execution (test without frying your neighbors' electronics)
- Production mode requires root privileges (because RF transmission isn't a joke)

## ⚡ Quick Start (Stop Reading, Start Transmitting)

```bash
go get github.com/psyb0t/gorpitx
```

```go
package main

import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

func main() {
    // Get the singleton instance (there can be only one)
    rpitx := gorpitx.GetInstance()

    // Configure PIFMRDS module (FM with RDS, fancy shit)
    args := map[string]interface{}{
        "freq":  107.9,  // MHz - pick a frequency, any frequency
        "audio": "/path/to/audio.wav",  // Your audio masterpiece
        "pi":    "1234",  // Station ID (4 hex digits)
        "ps":    "BADASS",  // Station name (8 chars max)
        "rt":    "Broadcasting from Go like a boss!",
    }

    argsJSON, _ := json.Marshal(args)
    ctx := context.Background()

    // Execute with timeout (because infinite loops are evil)
    err := rpitx.Exec(ctx, gorpitx.ModuleNamePIFMRDS, argsJSON, 5*time.Minute)
    if err != nil {
        panic(err)  // Shit hit the fan
    }
}
```

## 🔧 Installation Requirements (Don't Skip This Shit)

**Hardware**: Raspberry Pi with GPIO access (Pi Zero, Pi Zero W, Pi A+, Pi B+, Pi 2B, Pi 3B, Pi 3B+)
**OS**: Raspbian/Raspberry Pi OS (anything else is asking for trouble)
**Dependencies**: rpitx (install this beast first or nothing works)
**Privileges**: Must run as root in production (sudo your way to glory)

### Install rpitx (The Foundation of Everything)

```bash
# On your Pi, do this shit:
sudo apt update
git clone https://github.com/F5OEO/rpitx.git
cd rpitx
chmod +x install.sh
sudo ./install.sh  # This might take a hot minute
```

### Configure Path (Optional But Smart)

```bash
# Set rpitx binary path if you're not using defaults
export GORPITX_PATH="/home/pi/rpitx"
```

## 📋 PIFMRDS Module Configuration

```go
type PIFMRDS struct {
    Freq        float64  // Frequency in MHz (required, 0.005-1500 MHz)
    Audio       string   // Audio file path (required, must exist)
    PI          string   // PI code - 4 hex digits (optional)
    PS          string   // Station name - max 8 chars (optional)
    RT          string   // Radio text - max 64 chars (optional)
    PPM         *float64 // Clock correction ppm (optional)
    ControlPipe *string  // Named pipe for runtime control (optional)
}
```

**Validation Rules:**

- `Freq`: Required, positive, within RPiTX range (5kHz-1500MHz), 0.1MHz precision
- `Audio`: Required, file must exist (no stdin support yet)
- `PI`: Exactly 4 hexadecimal characters if specified
- `PS`: Max 8 characters, cannot be empty/whitespace if specified
- `RT`: Max 64 characters
- `ControlPipe`: Must exist if specified (create with `mkfifo`)

## 📻 TUNE Module Configuration

```go
type TUNE struct {
    Frequency     float64  // Hz, required, 50kHz-1500MHz
    ExitImmediate *bool    // Exit without killing carrier (optional)
    PPM           *float64 // Clock correction ppm > 0 (optional)
}
```

**Validation Rules:**

- `Frequency`: Required, positive, within RPiTX range (50kHz-1500MHz) in Hz
- `ExitImmediate`: Optional boolean, exits without killing carrier when true
- `PPM`: Optional, must be positive if specified

**Example Usage:**

```go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

args := gorpitx.TUNE{
    Frequency:     434000000.0,         // 434 MHz in Hz
    ExitImmediate: boolPtr(true),       // Exit without killing carrier
    PPM:           floatPtr(2.5),       // Clock correction
}

argsJSON, _ := json.Marshal(args)
ctx := context.Background()

// Execute carrier tune
err := rpitx.Exec(ctx, gorpitx.ModuleNameTUNE, argsJSON, 0) // No timeout
if err != nil {
    panic(err)
}

func floatPtr(f float64) *float64 { return &f }
func boolPtr(b bool) *bool { return &b }
```

## 🌊 PICHIRP Module Configuration

```go
type PICHIRP struct {
    Frequency float64 `json:"frequency"` // Hz, required, center frequency
    Bandwidth float64 `json:"bandwidth"` // Hz, required, sweep bandwidth
    Time float64 `json:"time"` // Seconds, required, sweep duration
}
```

**Validation Rules:**

- `Frequency`: Required, positive, within RPiTX range (50kHz-1500MHz) in Hz
- `Bandwidth`: Required, positive value in Hz
- `Time`: Required, positive value in seconds

**Example Usage:**

```go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

args := gorpitx.PICHIRP{
    Frequency: 434000000.0, // 434 MHz in Hz
    Bandwidth: 100000.0,    // 100 kHz bandwidth
    Time:      5.0,         // 5 seconds
}

argsJSON, _ := json.Marshal(args)
ctx := context.Background()

// Execute frequency sweep
err := rpitx.Exec(ctx, gorpitx.ModuleNamePICHIRP, argsJSON, 0) // No timeout
if err != nil {
    panic(err)
}
```

## ⚡ MORSE Module Configuration

```go
type MORSE struct {
    Frequency float64 `json:"frequency"` // Hz, required, carrier frequency
    Rate      int     `json:"rate"`      // Required, rate in dits per minute
    Message   string  `json:"message"`   // Required, message text to transmit
}
```

**Validation Rules:**

- `Frequency`: Required, positive, within RPiTX range (50kHz-1500MHz) in Hz
- `Rate`: Required, positive integer, dits per minute
- `Message`: Required, cannot be empty or whitespace only

**Example Usage:**

```go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

args := gorpitx.MORSE{
    Frequency: 14070000.0, // 14.070 MHz in Hz (popular CW frequency)
    Rate:      20,         // 20 dits per minute (standard speed)
    Message:   "CQ CQ DE N0CALL",
}

argsJSON, _ := json.Marshal(args)
ctx := context.Background()

// Execute morse transmission
err := rpitx.Exec(ctx, gorpitx.ModuleNameMORSE, argsJSON, 0) // No timeout
if err != nil {
    panic(err)
}
```

## 📟 POCSAG Module Configuration

```go
type POCSAG struct {
    Frequency float64 `json:"frequency"` // Hz, required, 50kHz-1500MHz
    BaudRate *int `json:"baudRate,omitempty"` // Optional, 512/1200/2400, default 1200
    FunctionBits *int `json:"functionBits,omitempty"` // Optional, 0-3, default 3
    NumericMode *bool `json:"numericMode,omitempty"` // Optional, default false
    RepeatCount *int `json:"repeatCount,omitempty"` // Optional, default 4
    InvertPolarity *bool `json:"invertPolarity,omitempty"` // Optional, default false
    Debug *bool `json:"debug,omitempty"` // Optional, default false
    Messages []POCSAGMessage `json:"messages"` // Required, address:message pairs
}

type POCSAGMessage struct {
    Address int `json:"address"` // Required, pager address
    Message string `json:"message"` // Required, message text
    FunctionBits *int `json:"functionBits,omitempty"` // Optional override
}
```

**POCSAG Stdin Implementation:**

POCSAG uses **stdin for message data** (like the native rpitx binary), not command arguments. Messages are automatically formatted as `address:message` pairs separated by newlines and sent via stdin to the rpitx POCSAG binary.

**Validation Rules:**

- `Frequency`: Required, positive, within RPiTX range (50kHz-1500MHz) in Hz
- `BaudRate`: Optional, must be 512, 1200, or 2400
- `FunctionBits`: Optional, must be 0-3
- `NumericMode`: Optional boolean flag for numeric mode
- `RepeatCount`: Optional, must be positive
- `InvertPolarity`: Optional boolean flag to invert polarity
- `Debug`: Optional boolean flag for debug mode
- `Messages`: Required slice with at least one message

**Note**: Optional parameters use rpitx defaults if not specified (1200 baud, function bits 3, repeat count 4). Frequency is required at the gorpitx level for validation.

**How POCSAG Stdin Works:**

The rpitx POCSAG binary expects message data via stdin in `address:message` format:

```bash
# Native rpitx usage:
printf "123456:Emergency alert\n789012:Second message" | sudo ./pocsag -f 466230000 -r 1200
```

GoRPITX automatically handles this:

1. Extracts messages from JSON configuration
2. Formats as `address:message` pairs with newline separation
3. Sends via stdin to the rpitx POCSAG binary
4. Command arguments contain only flags (`-f`, `-r`, etc.)

**Example Usage:**

```go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

args := gorpitx.POCSAG{
    Frequency:      466230000.0,           // 466.230 MHz (pager frequency)
    BaudRate:       intPtr(1200),          // 1200 baud
    FunctionBits:   intPtr(3),             // Function 3
    NumericMode:    boolPtr(false),        // Text mode
    RepeatCount:    intPtr(4),             // Repeat 4 times
    InvertPolarity: boolPtr(false),        // Normal polarity
    Debug:          boolPtr(false),        // No debug
    Messages: []gorpitx.POCSAGMessage{
        {
            Address: 123456,
            Message: "Emergency alert test message",
        },
        {
            Address: 789012,
            Message: "Second pager message",
        },
    },
}

argsJSON, _ := json.Marshal(args)
ctx := context.Background()

// Execute POCSAG transmission
// Equivalent to: printf "123456:Emergency alert test message\n789012:Second pager message" | sudo ./pocsag -f 466230000 -r 1200 -b 3 -t 4
err := rpitx.Exec(ctx, gorpitx.ModuleNamePOCSAG, argsJSON, 0) // No timeout
if err != nil {
    panic(err)
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }
func boolPtr(b bool) *bool { return &b }
```

## 🎨 SPECTRUMPAINT Module Configuration

```go
type SPECTRUMPAINT struct {
    PictureFile string   `json:"pictureFile"` // Required, path to raw data file
    Frequency   float64  `json:"frequency"`   // Hz, required, carrier frequency
    Excursion   *float64 `json:"excursion,omitempty"` // Hz, optional, frequency excursion
}
```

**Validation Rules:**

- `PictureFile`: Required, file must exist (expects raw YUV data format, 320 pixels wide)
- `Frequency`: Required, positive, within RPiTX range (50kHz-1500MHz) in Hz
- `Excursion`: Optional, must be positive if specified

**Image Format Requirements:**
The spectrumpaint binary expects raw YUV data files with a fixed width of 320 pixels. Convert your images using ImageMagick:

```bash
# Convert any image to the required 320-pixel wide format (creates multiple files: picture.Y, picture.U, picture.V)
convert input.jpg -resize 320x -flip -quantize YUV -dither FloydSteinberg -colors 4 -interlace partition picture.yuv

# The spectrumpaint binary uses the luminance channel (picture.Y file)
# For specific height (e.g., 100 pixels):
convert input.jpg -resize 320x100! -flip -quantize YUV -dither FloydSteinberg -colors 4 -interlace partition picture.yuv
```

**Important**: The `-interlace partition` option creates separate Y, U, V files. Use the `.Y` file (luminance channel) with spectrumpaint.

**Note**: The 320-pixel width limit is hardcoded in the rpitx spectrumpaint binary.

**Example Usage:**

```go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

args := gorpitx.SPECTRUMPAINT{
    PictureFile: ".fixtures/test_spectrum_320x100.Y", // YUV luminance file
    Frequency:   434000000.0, // 434 MHz in Hz
    Excursion:   floatPtr(100000.0), // 100 kHz excursion
}

argsJSON, _ := json.Marshal(args)
ctx := context.Background()

// Execute spectrum paint transmission
err := rpitx.Exec(ctx, gorpitx.ModuleNameSPECTRUMPAINT, argsJSON, 0) // No timeout
if err != nil {
    panic(err)
}

func floatPtr(f float64) *float64 { return &f }
```

## 📡 FT8 Module Configuration

```go
type FT8 struct {
    Frequency float64  `json:"frequency"`           // Hz, required, carrier frequency
    Message   string   `json:"message"`             // Required, FT8 message
    PPM       *float64 `json:"ppm,omitempty"`       // Optional, clock correction ppm
    Offset    *float64 `json:"offset,omitempty"`    // Hz, optional, frequency offset (0-2500)
    Slot      *int     `json:"slot,omitempty"`      // Optional, time slot 0/1/2
    Repeat    *bool    `json:"repeat,omitempty"`    // Optional, repeat mode (every 15s)
}
```

**Validation Rules:**

- `Frequency`: Required, positive, within RPiTX range (50kHz-1500MHz) in Hz
- `Message`: Required, cannot be empty/whitespace
- `PPM`: Optional, clock correction value (positive, negative, or zero)
- `Offset`: Optional, frequency offset 0-2500 Hz (default: 1240 Hz if not specified)
- `Slot`: Optional, time slot: 0 (first 15s), 1 (second 15s), 2 (always/every 15s)
- `Repeat`: Optional, enables repeat mode (transmit every 15 seconds)

**FT8 Protocol Details:**

FT8 is a weak-signal digital mode designed for amateur radio communication. Key characteristics:
- 15-second transmission periods with precise timing
- Uses 8-FSK modulation with 6.25 Hz tone spacing
- Default frequency offset of 1240 Hz within the FT8 sub-band
- Message length handled by the pift8 binary

**Example Usage:**

```go
import (
    "context"
    "encoding/json"
    "time"
    "github.com/psyb0t/gorpitx"
)

// Basic CQ call
args := gorpitx.FT8{
    Frequency: 14074000.0,    // 14.074 MHz (20m FT8 frequency)
    Message:   "CQ W1AW FN31", // Standard FT8 CQ format
}

argsJSON, _ := json.Marshal(args)
ctx := context.Background()

// Execute single FT8 transmission
err := rpitx.Exec(ctx, gorpitx.ModuleNameFT8, argsJSON, 0) // No timeout
if err != nil {
    panic(err)
}

// Advanced configuration with repeat mode
advancedArgs := gorpitx.FT8{
    Frequency: 7074000.0,             // 7.074 MHz (40m FT8 frequency)
    Message:   "K0HAM W5XYZ",         // Directed call
    PPM:       floatPtr(2.5),         // Clock correction
    Offset:    floatPtr(1500.0),      // Custom offset frequency
    Slot:      intPtr(1),             // Second time slot
    Repeat:    boolPtr(true),         // Repeat every 15 seconds
}

argsJSON2, _ := json.Marshal(advancedArgs)

// Execute repeating FT8 transmission (press Ctrl+C to stop)
err = rpitx.Exec(ctx, gorpitx.ModuleNameFT8, argsJSON2, 0)
if err != nil {
    panic(err)
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }
func boolPtr(b bool) *bool { return &b }
```

**Common FT8 Frequencies:**

- **20m**: 14.074 MHz
- **40m**: 7.074 MHz
- **80m**: 3.573 MHz
- **15m**: 21.074 MHz
- **10m**: 28.074 MHz

**Message Format Examples (Typical QSO Sequence):**

1. CQ call: `CQ W1AW FN31` (callsign + grid square)
2. Reply: `W1AW K0HAM EM69` (their call + your call + your grid)
3. Signal report: `K0HAM W1AW -15` (signal strength in dB)
4. Report + confirm: `W1AW K0HAM R-08` (R = received, your report)
5. Nearly complete: `K0HAM W1AW RR73` (RR = received + 73)
6. QSO complete: `W1AW K0HAM 73` (final acknowledgment)

**Other Valid Formats:**
- Contest exchanges: `W1AW K0HAM 599 CA` (RST + state/province)
- DX calls: `CQ DX K0HAM EM69`
- Directed CQ: `CQ NA W1AW FN31` (North America only)

**Note**: All FT8 operations should follow amateur radio band plans and regulations. Use appropriate power levels and ensure proper station identification.

## 🎛️ Process Control

### Stream Output

**Option 1: Async Streaming (Recommended)**

```go
stdout := make(chan string, 100)
stderr := make(chan string, 100)

// Start async streaming (can be called before execution)
rpitx.StreamOutputsAsync(stdout, stderr)

// Start output collection
go func() {
    for line := range stdout {
        fmt.Println("STDOUT:", line)
    }
}()

// Execute - streaming will automatically start when execution begins
err := rpitx.Exec(ctx, gorpitx.ModuleNameMORSE, argsJSON, 30*time.Second)
```

**Option 2: Manual Streaming (Requires precise timing)**

```go
stdout := make(chan string, 100)
stderr := make(chan string, 100)

// Start execution in goroutine
go func() {
    err := rpitx.Exec(ctx, gorpitx.ModuleNameMORSE, argsJSON, 30*time.Second)
    // Handle error
}()

// Wait for execution to start, then stream
time.Sleep(100 * time.Millisecond)
rpitx.StreamOutputs(stdout, stderr)  // Only works during execution

go func() {
    for line := range stdout {
        fmt.Println("STDOUT:", line)
    }
}()
```

### Graceful Stop

```go
ctx := context.Background()
err := rpitx.Stop(ctx, 3*time.Second)
if err != nil {
    // Handle stop error
}
```

### Execution State

- Only one module can execute at a time
- `Exec()` blocks until completion or timeout
- Automatic cleanup on context cancellation
- Process termination with SIGTERM then SIGKILL

## ⚙️ Environment Configuration

### Development Mode

Set `ENV=dev` to enable mock execution:

```bash
ENV=dev go run main.go
```

Mock execution runs infinite loop printing status every second instead of actual RF transmission.

### Production Mode

Default mode requiring root privileges:

```bash
sudo go run main.go  # or deploy as root
```

Executes actual rpitx binaries with proper RF transmission.

## 🧪 Error Handling

**Module Errors:**

- `ErrUnknownModule`: Requested module not registered
- `ErrExecuting`: Another command already running
- `ErrNotExecuting`: No active execution for stop/stream

**Validation Errors:**

- `commonerrors.ErrRequiredFieldNotSet` - Missing required fields (wrapped with field name)
- `commonerrors.ErrInvalidValue` - Invalid parameter values (wrapped with details)
- `commonerrors.ErrFileNotFound` - Missing files (wrapped with file path)
- `ErrFreqOutOfRange`, `ErrFreqPrecision` - Frequency validation errors
- `ErrPIInvalidHex` - PI code validation
- `ErrPSTooLong` - PS text validation

**Note**: All validation errors use `ctxerrors.Wrap()` pattern for contextual error information.

## 🔗 Architecture

### Module Interface

```go
type Module interface {
    ParseArgs(json.RawMessage) ([]string, io.Reader, error)
}
```

New modules implement this interface with:

1. JSON unmarshaling of configuration
2. Parameter validation
3. Command-line argument building
4. Stdin data preparation (return `nil` if no stdin needed)

**Stdin Usage:**

- Most modules return `nil` for stdin (TUNE, MORSE, PIFMRDS, PICHIRP, SPECTRUMPAINT)
- POCSAG returns `io.Reader` with message data in `address:message` format
- Commander automatically pipes stdin data to the rpitx binary when provided

### Frequency Utilities

- `hzToMHz(hz float64) float64` - Convert Hz to MHz
- `mHzToHz(mHz float64) float64` - Convert MHz to Hz
- `kHzToMHz(kHz float64) float64` - Convert kHz to MHz
- `mHzToKHz(mHz float64) float64` - Convert MHz to kHz
- `isValidFreqHz(freqHz float64) bool` - Validate Hz frequency (standardized)
- `getMinFreqMHzDisplay() float64` - Get min frequency for error displays (0.005 MHz)
- `getMaxFreqMHzDisplay() float64` - Get max frequency for error displays (1500 MHz)
- `hasValidFreqPrecision(freqMHz float64) bool` - Check 0.1MHz precision

**Note**: pifmrds uses MHz, other planned modules use Hz.

## 📋 TODO: Remaining Modules Implementation (The Fun Stuff)

Based on the easytest modules from rpitx, here are the **5 badass modules** we still need to implement (excluding that legacy rpitx garbage):

- **SENDIQ** - IQ Data Transmission

  - **Command**: `sendiq [-i File] [-s Samplerate] [-f Frequency] [-l] [-h Harmonic] [-m Token] [-d] [-p Power] [-t IQType]`
  - **Go struct**:
    ```go
    type SENDIQ struct {
        InputFile string `json:"inputFile"` // Required, input file path
        SampleRate *int `json:"sampleRate,omitempty"` // Optional, 10000-250000, default 48000
        Frequency *float64 `json:"frequency,omitempty"` // Hz, optional, 50kHz-1500MHz, default 434e6
        LoopMode *bool `json:"loopMode,omitempty"` // Optional, default false
        Harmonic *int `json:"harmonic,omitempty"` // Optional, >= 1, default 1
        SharedMemoryToken *int `json:"sharedMemoryToken,omitempty"` // Optional
        DDSMode *bool `json:"ddsMode,omitempty"` // Optional, default false
        PowerLevel *float64 `json:"powerLevel,omitempty"` // Optional, 0.0-7.0, default 0.1
        IQType *string `json:"iqType,omitempty"` // Optional, i16/u8/float/double, default "i16"
    }
    ```
  - **Validation**: File exists, sample rate 10000-250000, power 0.0-7.0, IQ type enum

- **FREEDV** - FreeDV Digital Voice

  - **Command**: `freedv vco.rf frequency(Hz) [samplerate(Hz)]`
  - **Go struct**:
    ```go
    type FREEDV struct {
        VCOFile string `json:"vcoFile"` // Required, .rf file path
        Frequency float64 `json:"frequency"` // Hz, required
        SampleRate *int `json:"sampleRate,omitempty"` // Hz, optional, default 400
    }
    ```
  - **Validation**: File exists, frequency > 0, sample rate > 0 if provided

- **PISSTV** - Slow Scan TV

  - **Command**: `pisstv picture.rgb frequency(Hz)`
  - **Go struct**:
    ```go
    type PISSTV struct {
        PictureFile string `json:"pictureFile"` // Required, 320x256 .rgb file
        Frequency float64 `json:"frequency"` // Hz, required
    }
    ```
  - **Validation**: File exists, frequency > 0

- **PIOPERA** - OPERA Protocol

  - **Command**: `piopera CALLSIGN OperaMode[0.5,1,2,4,8] frequency(Hz)`
  - **Go struct**:
    ```go
    type PIOPERA struct {
        Callsign string `json:"callsign"` // Required, amateur radio callsign
        Mode float64 `json:"mode"` // Required, 0.5/1/2/4/8
        Frequency float64 `json:"frequency"` // Hz, required
    }
    ```
  - **Validation**: Callsign format (3rd char numeric), mode enum, frequency > 0

- **PIRTTY** - RTTY Protocol
  - **Command**: `pirtty frequency(Hz) SpaceFrequency(Hz) text`
  - **Go struct**:
    ```go
    type PIRTTY struct {
        Frequency float64 `json:"frequency"` // Hz, required, carrier frequency
        SpaceFrequency int `json:"spaceFrequency"` // Hz, required, space tone
        Text string `json:"text"` // Required, message text
    }
    ```
  - **Validation**: All required, frequency > 0, space frequency > 0, text not empty

### Common Validation Functions Needed

```go
func ValidateFrequency(freq float64, min, max float64) error
func ValidateFileExists(path string) error
func ValidateEnum(value string, allowedValues []string) error
func ValidateRange(value, min, max float64) error
```

## ⚠️ Legal Notice (Read This or Get Fucked by the FCC)

**RF transmission is regulated as hell.** Don't be a dickhead - get proper licensing before broadcasting. This software is for:

- Licensed amateur radio operators (you know who you are)
- Low-power experimentation in permitted bands (don't fry your neighbor's radio gear)
- Educational/research purposes (learn responsibly, you beautiful bastards)

**Absolutely NOT for**: Commercial broadcasting without authorization (the FCC will skull-fuck your wallet).

## 📚 Dependencies

- [`github.com/psyb0t/commander`](https://github.com/psyb0t/commander) - Process execution
- [`github.com/psyb0t/common-go/env`](https://github.com/psyb0t/common-go) - Environment detection
- [`github.com/psyb0t/ctxerrors`](https://github.com/psyb0t/ctxerrors) - Context-aware errors
- [`github.com/psyb0t/gonfiguration`](https://github.com/psyb0t/gonfiguration) - Configuration parsing
- [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus) - Logging

## 📄 License

MIT License. Use responsibly and don't be a twat.

---

_Go interface for rpitx that doesn't suck. Built for radio enthusiasts who want clean code without the usual C library nightmare fuel._
