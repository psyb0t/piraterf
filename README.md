# 🏴‍☠️ PIrateRF - Software-Defined Radio Transmission Platform

**PIrateRF** is a fucking badass software-defined radio (SDR) transmission platform that turns your **Raspberry Pi Zero W** into a portable RF signal generator with a sleek web interface. This beast enables you to transmit various types of radio signals including FM radio broadcasts, Morse code, carrier waves, and even spectrum painting - all controlled through your browser like a proper pirate! 📡⚡

## 📋 Table of Contents

- [What the Fuck Does This Thing Do?](#-what-the-fuck-does-this-thing-do)
- [Architecture & Technology Stack](#️-architecture--technology-stack)
- [Quick Fucking Setup](#-quick-fucking-setup)
- [Development Workflow](#️-development-workflow)
- [RF Transmission Modes](#-rf-transmission-modes)
- [Network Configuration](#-network-configuration)
- [Project Structure](#-project-structure)
- [Make Targets Reference](#-make-targets-reference)
- [Configuration](#-configuration)
- [Audio Processing Pipeline](#-audio-processing-pipeline)
- [Image Processing for Spectrum Painting](#️-image-processing-for-spectrum-painting)
- [Legal and Safety Notice](#️-legal-and-safety-notice)
- [Contributing](#-contributing)
- [License](#-license)
- [Dependencies](#-dependencies)

## 🎯 What the Fuck Does This Thing Do?

PIrateRF transforms your Pi Zero into a **standalone RF transmission station** that can:

- **🎵 FM Radio Broadcasting**: Transmit FM modulated audio with full RDS (Radio Data System) metadata including station names, radio text, and PI codes
- **📻 Morse Code Transmission**: Send CW (continuous wave) Morse code signals
- **🎛️ Carrier Wave Generation**: Simple carrier frequency generation for testing and tuning
- **🌊 Frequency Sweep**: Generate carrier frequency sweeps for RF testing and analysis or just for teh lulz
- **📟 POCSAG Paging**: Transmit POCSAG pager messages with configurable baud rates, function bits, and multi-message support
- **📡 FT8 Digital Mode**: Extreme long-range digital amateur radio protocol capable of intercontinental communication with minimal power
- **📺 PISSTV (SSTV)**: Slow Scan Television transmission using Martin 1 protocol for sending images over amateur radio frequencies
- **📠 PIRTTY (RTTY)**: Radio Teletype transmission using Baudot code and frequency shift keying for text communication
- **🎨 Spectrum Painting**: Transmit images as RF spectrum patterns (because why the fuck not?)
- **🎧 Audio Processing**: Upload files or record via microphone through the browser
- **📱 Web-based Control**: Full-featured HTML5 interface with live WebSocket updates

All of this runs on a **Pi Zero W configured as a WiFi access point**, making it a completely standalone, portable RF transmission platform that you can take anywhere and control from any device with a browser.

## 🏗️ Architecture & Technology Stack

### Core Components

- **Backend**: Go 1.24.6 with ARM cross-compilation for Pi Zero efficiency
- **RF Engine**: [`gorpitx`](https://github.com/psyb0t/gorpitx) - Go wrapper for the legendary [rpitx](https://github.com/F5OEO/rpitx) C library
- **Web Framework**: Custom HTTP server with WebSocket support via [`aichteeteapee`](https://github.com/psyb0t/aichteeteapee)
- **Frontend**: Modern HTML5/CSS3/JavaScript with real-time communication
- **Audio Processing**: Sox and FFmpeg for professional audio conversion and manipulation
- **Service Framework**: Custom [`servicepack`](https://github.com/psyb0t/servicepack) framework for project structure and deployment

### Service Architecture

PIrateRF is a **single Go service** with modular components:

- **RF Transmission Engine**: Core logic for generating FM, Morse, and spectrum signals
- **Execution Manager**: Handles RF transmission execution with atomic state control preventing concurrent transmissions
- **WebSocket Hub**: Real-time bidirectional communication with the frontend interface
- **HTTP Server**: Serves the web interface and handles secure file uploads
- **Audio/Image Processing**: Automatic format conversion and optimization pipelines

## 🚀 Quick Fucking Setup

### Prerequisites

- **Raspberry Pi Zero W**
- **SD Card** (8GB+ recommended)
- **Docker** for development
- Basic knowledge of RF regulations in your area (don't be a fucking idiot)

### 🚨 IMPORTANT: Pi Zero Setup First!

**Before you do ANYTHING else**, you need to get your fucking Pi Zero W connected and accessible via SSH. Follow this tutorial that actually fucking works:

👉 **[Pi Zero W USB Connection Tutorial](https://ciprian.51k.eu/pi-zero-1-wh-ubuntu-24-04-usb-connection-the-tutorial-that-actually-fkin-works/)**

This will get your Pi Zero connected via USB with SSH access so you can actually deploy PIrateRF to the bastard. Don't skip this step or you'll be fucked trying to connect to your Pi later!

**Credentials setup**: During the Pi Zero setup tutorial, I used username `fucker`, hostname `piraterf.local`, password `FUCKER`. Use these same credentials if you don't want to get confused with the rest of this tutorial - otherwise go ahead and set up some lame ass shit with your own credentials.

**🌐 INTERNET SHARING REQUIRED**: After USB connection is working, you MUST share internet from your computer to the Pi Zero. The setup scripts need to download packages and dependencies. Your Pi Zero connects via USB but has no internet unless you share it from your host computer.

**Set up internet sharing on Ubuntu/Linux** (complete fucking setup):

1. **Set connection to shared**: In Ubuntu Network Settings, find the USB connection (usually shows as `usb0` or similar), click on it, go to IPv4 settings, and change the method from "Link-Local Only" to "Shared to other computers". This is fucking critical!

2. **Stop Docker services** (they fuck with networking):

```bash
sudo systemctl stop docker.socket
sudo systemctl stop docker
```

3. **Restart NetworkManager**:

```bash
sudo systemctl restart NetworkManager
```

4. **Configure iptables rules** (replace `usb0` and `enp5s0` with your actual interfaces):

```bash
sudo iptables -A INPUT -i lo -j ACCEPT
sudo iptables -A OUTPUT -o lo -j ACCEPT
sudo iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
sudo iptables -A FORWARD -m state --state ESTABLISHED,RELATED -j ACCEPT
sudo iptables -A FORWARD -i usb0 -o enp5s0 -j ACCEPT
sudo iptables -A FORWARD -i enp5s0 -o usb0 -j ACCEPT
sudo iptables -t nat -A POSTROUTING -s 10.42.0.0/24 -o enp5s0 -j MASQUERADE
sudo iptables -P INPUT ACCEPT
sudo iptables -P FORWARD ACCEPT
sudo iptables -P OUTPUT ACCEPT
```

5. **Test it**: SSH into your Pi and run `ping 8.8.8.8` - if it works, you're fucking golden!

**Find your interfaces**: Use `ip link show` to see `usb0` (Pi connection) and your main internet interface (usually `eth0`, `wlan0`, `enp0s3`, etc.). Without this internet sharing setup, the dependency installation will fail because the Pi can't reach package repositories!

### 🔌 ANTENNA SETUP

Connect your fucking antenna to the Pi Zero:

**GPIO Connection**: Connect your antenna cable to **GPIO 4 (Physical Pin 7)** on the Pi Zero W. This is your RF output pin.

**Antenna Options**:

1. **Short wire** (~10-20cm): Best for indoor testing and learning. Keeps power low and legal.
2. **75cm wire**: Longer range but too much signal leaks outside your property - use this **ONLY indoors**
3. **Proper antenna with low pass filter**: Build or buy a proper antenna system with SMA connector and appropriate low pass filter for your frequency

**⚠️ IMPORTANT**: The 75cm antenna is a fucking liability outdoors - your signal will travel way beyond your property and you'll be violating regulations. Stick to short wires for testing or build proper filtered antenna systems for serious use.

**⚠️ USE A FUCKING LOW PASS FILTER!** The Pi GPIO outputs square waves which generate harmonics across the entire spectrum. Without proper filtering, you'll spray RF energy all over the fucking place and violate spurious emission regulations. Always use an appropriate low pass filter for your transmission frequency!

### 1. Initial Pi Setup and Configuration

Flash Raspberry Pi OS Lite to your SD card and enable SSH. Then:

```bash
# Clone this badass project
git clone https://github.com/psyb0t/piraterf.git
cd piraterf
```

Edit scripts/pi_config.sh and modify these values to match your Pi:

```bash
export PI_USER="fucker"              # Pi username
export PI_HOST="piraterf.local"      # Pi hostname/IP
export PI_PASS="FUCKER"             # Pi password

# WiFi AP Configuration
export AP_SSID="🏴‍☠️📡"             # WiFi access point name
export AP_PASSWORD="FUCKER!!!"      # WiFi access point password
export AP_CHANNEL="7"                # WiFi channel (1-14)
export AP_COUNTRY="US"               # WiFi country code (US, UK, DE, etc.)
```

### 2. Complete Automated Setup

Run the full setup pipeline that configures everything automatically:

```bash
make pi
```

This fucking command will:

1. **Install dependencies** (rpitx, sox, ffmpeg, etc.)
2. **Configure WiFi Access Point** (SSID: "🏴‍☠️📡", Password: "FUCKER!!!")
3. **Setup system branding** (MOTD, terminal aliases, pirate theme)
4. **Build and deploy** the PIrateRF application
5. **Install systemd service** for auto-start
6. **Reboot** the Pi into pirate mode

### 3. Connect and Use

After reboot:

1. **Connect to WiFi**: "🏴‍☠️📡" with password "FUCKER!!!"
2. **Open browser**: Navigate to `https://piraterf.local` (or whatever hostname you configured)
3. **Start transmitting**: Upload audio, configure RDS, and broadcast like a proper pirate!

**⚠️ IMPORTANT**: Use the HTTPS hostname for full functionality. The fucking microphone recording feature requires HTTPS with a proper hostname to work due to browser security restrictions.

**🎉 Pirate Crew Mode**: Connect multiple devices to the same WiFi network and all access the web interface simultaneously! While only one transmission can run at a time (because GPIO doesn't fucking share), all connected devices see real-time transmission status, output logs, and can take turns controlling the RF transmissions. Perfect for fucking around with friends in a radio wave gangbang! 📡💥

## 🛠️ Development Workflow

### Local Development

```bash
# Run locally in development mode
make run-dev

# Format and lint code
make lint-fix

# Run tests with coverage
make test-coverage

# Build for production
make build
```

### Pi Development Cycle

```bash
# Cross-compile for Pi
make build

# Deploy to Pi
make deploy

# Install service
make install

# SSH into Pi for debugging
make ssh

# View logs
make ssh
# Then: sudo journalctl -fu piraterf
```

### Individual Pi Setup Commands

If you want to run setup steps individually:

```bash
make pi-setup-deps      # Install rpitx and dependencies
make pi-setup-ap        # Configure Pi as WiFi access point
make pi-setup-branding  # System branding setup
make deploy             # Copy files to Pi
make install            # Install as systemd service
make pi-reboot          # Reboot Pi
make uninstall          # Remove PIrateRF from Pi
```

## 📡 RF Transmission Modes

### FM Radio Broadcasting

- **Audio Support**: MP3, WAV, and more (auto-converted to 48kHz/16-bit/mono)
- **RDS Metadata**: Station name, radio text, and PI codes
- **Playlist Mode**: Create playlists with intro/outro support
- **Play Once Mode**: Single play with automatic timeout calculation
- **Microphone Recording**: Record audio directly through the browser interface

### Morse Code (CW)

- **Text to Morse**: Automatic conversion with configurable WPM (words per minute)
- **CW Transmission**: Continuous wave RF transmission

### Tune Mode

- **Carrier Wave**: Simple carrier wave generation for testing and verification

### Frequency Sweep (CHIRP)

- **Configurable Sweeps**: Generate carrier wave sweeps with customizable center frequency, bandwidth, and duration
- **RF Testing**: Perfect for antenna analysis, filter characterization, and RF circuit testing
- **Wide Range Support**: Supports frequencies from 50 kHz to 1500 MHz with variable bandwidth

### POCSAG Paging

- **Message Transmission**: Send pager messages with configurable addresses and text content
- **Multiple Messages**: Support for multiple messages in a single transmission
- **Configurable Parameters**: Adjust baud rate (512, 1200, 2400 bps), function bits, and transmission options
- **Flexible Options**: Numeric mode, polarity inversion, and debug mode with toggle controls

### FT8 Digital Mode

- **Extreme Long-Range Communication**: Digital protocol designed for intercontinental communication with proper amplification and antenna systems
- **Precise Timing**: 15-second transmission periods with automated time slot control (0/1/2)
- **8-FSK Modulation**: Uses 8-level frequency-shift keying with 6.25 Hz tone spacing
- **Frequency Management**: Configurable frequency offset (0-2500 Hz, default 1240 Hz) within FT8 sub-bands
- **Repeat Mode**: Optional continuous transmission every 15 seconds for beacon operation
- **Clock Correction**: PPM correction support for frequency accuracy
- **Common Frequencies**: 20m (14.074), 40m (7.074), 80m (3.573), 15m (21.074), 10m (28.074) MHz
- **Message Format**: Standard FT8 exchange formats (CQ calls, signal reports, grid squares, contest exchanges)

### PISSTV (Slow Scan Television)

- **Amateur Radio Image Transmission**: Send dick pics over radio using audio frequency modulation
- **Martin 1 Protocol**: Industry-standard SSTV protocol with automatic VIS header identification
- **RGB Image Support**: Upload images that are automatically converted to 320x256 RGB format
- **Compatible Reception**: Works with QSSTV (Linux), MMSSTV (Windows), Robot36 (Android), and other SSTV software
- **Common SSTV Frequencies**: 2m band (144.500 MHz), 70cm band (434.000 MHz), and other amateur allocations
- **Format Support**: JPEG, PNG, GIF with automatic RGB conversion for transmission

### PIRTTY (Radio Teletype)

- **Digital Text Communication**: Transmit text messages using Baudot code and frequency shift keying
- **FSK Modulation**: Standard 170 Hz shift between mark and space frequencies for amateur radio compatibility
- **Baudot Character Set**: 5-bit encoding with automatic LTRS/FIGS mode switching for letters, numbers, and symbols
- **Configurable Frequencies**: User-defined space frequency with automatic mark frequency calculation (space + 170 Hz)
- **Standard Baud Rate**: 45.45 baud (22ms per bit) following amateur radio RTTY conventions
- **Common RTTY Frequencies**: 20m (14.080-14.099), 40m (7.035-7.045), 80m (3.580-3.600 MHz)
- **Compatible Reception**: Works with standard RTTY software and terminal units

### Spectrum Painting

- **Image Upload**: Convert images to RF spectrum patterns
- **Format Support**: JPEG, PNG, GIF with automatic YUV conversion
- **Visual RF**: Turn your images into radio art (because pirates love art too)

## 🌐 Network Configuration

The `make pi-setup-ap` command configures the Pi as a **standalone WiFi access point**. Access point settings (SSID, password, IP ranges, etc.) are configured in `scripts/make/pi_setup_ap.sh`.

## 📁 Project Structure

```
piraterf/
├── cmd/                    # Main application entry points
├── internal/pkg/services/
│   └── piraterf/          # Core PIrateRF service implementation
│       ├── piraterf.go    # Main service logic
│       ├── http_server.go # Web server and API
│       ├── websocket*.go  # Real-time communication
│       ├── audio_*.go     # Audio processing pipeline
│       ├── image_*.go     # Image processing for spectrum paint
│       └── execution_*.go # RF transmission management
├── scripts/
│   ├── pi_config.sh       # Central configuration for Pi setup
│   ├── piraterf.sh        # Main Pi runtime script
│   ├── setup_*.sh         # Pi setup scripts (deps, AP, branding)
│   ├── deploy.sh          # Pi deployment script
│   ├── install.sh         # Pi service installation
│   ├── uninstall.sh       # Pi service removal
│   └── make/              # Make target implementations
│       ├── pi.sh          # Complete Pi setup pipeline
│       ├── build.sh       # Cross-compilation for Pi
│       ├── deploy.sh      # Deployment automation
│       ├── install.sh     # Service installation
│       ├── pi_setup_*.sh  # Pi configuration scripts
│       ├── ssh.sh         # SSH connection helper
│       ├── tls.sh         # TLS certificate generation
│       └── servicepack/   # Framework scripts
├── html/                  # Web interface templates
├── static/                # CSS, JavaScript, images
├── files/                 # Audio and image file storage
├── uploads/               # Temporary upload staging
├── Makefile              # Main build configuration
└── Makefile.servicepack  # Framework integration
```

## 🧰 Make Targets Reference

### Development

- `make run-dev` - Run locally with development setup
- `make build` - Cross-compile for ARM/Pi Zero
- `make lint-fix` - Format code and fix linting issues
- `make test-coverage` - Run tests with coverage analysis

### Pi Management

- `make pi-setup-deps` - Install rpitx and system dependencies
- `make pi-setup-ap` - Configure WiFi access point
- `make pi-setup-branding` - Setup system branding and accounts
- `make deploy` - Copy built files to Pi
- `make install` - Install and start systemd service
- `make pi` - Run full setup pipeline
- `make ssh` - SSH into the Pi
- `make pi-reboot` - Reboot the Pi
- `make uninstall` - Remove PIrateRF from Pi

### Utilities

- `make tls` - Generate TLS certificates
- `make clean` - Clean build artifacts
- `make help` - Show all available targets

## 🔧 Configuration

### Pi Connection Settings

All Pi configuration values are centralized in `scripts/pi_config.sh`:

- **PI_USER, PI_HOST, PI_PASS**: SSH connection credentials
- **AP_SSID, AP_PASSWORD**: WiFi access point credentials
- **AP_CHANNEL**: WiFi channel (1-14, avoid crowded channels)
- **AP_COUNTRY**: WiFi country code for regulatory compliance

### Service Configuration

The PIrateRF service uses environment variables for configuration. See `scripts/piraterf.sh` for all available configuration options and their default values.

## 🎵 Audio Processing Pipeline

PIrateRF automatically processes uploaded audio through a sophisticated pipeline:

1. **Format Detection**: Supports MP3, WAV, FLAC, OGG, and more
2. **Conversion**: Automatically converts to 48kHz, 16-bit, mono WAV using Sox
3. **Validation**: Ensures audio meets RF transmission requirements
4. **Storage**: Organizes files in `/files/audio/uploads/` and `/files/audio/sfx/`
5. **Playlist Support**: Create playlists with intro/outro and repeat modes

## 🖼️ Image Processing for Spectrum Painting & SSTV

Images are processed for RF transmission in both Spectrum Painting and PISSTV modes:

1. **Format Support**: JPEG, PNG, GIF automatically detected
2. **Dual Conversion**:
   - **YUV format (.Y files)**: For Spectrum Painting RF transmission
   - **RGB format (.rgb files)**: For PISSTV/SSTV transmission (320x256 resolution)
3. **Optimization**: Automatically resized and optimized for each transmission mode
4. **Storage**: Organized in `/files/images/uploads/` with both .Y and .rgb versions created

## 🏴‍☠️ Legal and Safety Notice

**IMPORTANT**: This software enables RF transmission. **You are responsible for complying with your local RF regulations and licensing requirements.**

- Ensure you have proper licenses for your transmission frequency and power levels
- Some frequencies require amateur radio licenses
- Respect power limitations and spurious emission requirements
- Don't interfere with emergency services or licensed operators
- When in doubt, consult your local RF regulatory authority

**⚠️ USE A FUCKING LOW PASS FILTER!** The Pi GPIO outputs square waves which generate harmonics across the entire spectrum. Without proper filtering, you'll spray RF energy all over the fucking place and violate spurious emission regulations. Always use an appropriate low pass filter for your transmission frequency!

**PIrateRF is designed for educational, experimental, and licensed amateur radio use. The developers are not responsible for any misuse or regulatory violations.**

## 🤝 Contributing

Want to make this pirate ship even more badass?

1. Fork the repository
2. Create a feature branch
3. Make your fucking awesome changes
4. Test on actual Pi hardware
5. Submit a pull request with a clear description

### Development Guidelines

- Follow the existing code style (use `make lint-fix`)
- Write tests for new features
- Update documentation for any new functionality
- Test on real Pi Zero W hardware before submitting

## 📝 License

This project is licensed under WTFPL (Do What The Fuck You Want To Public License).

## 🔗 Dependencies

- **[rpitx](https://github.com/F5OEO/rpitx)** - The legendary RF transmission library that makes this all possible
- **[servicepack](https://github.com/psyb0t/servicepack)** - The framework that keeps this project organized and deployable
- **[aichteeteapee](https://github.com/psyb0t/aichteeteapee)** - The HTTP server framework powering the web interface
- **[gorpitx](https://github.com/psyb0t/gorpitx)** - Go wrapper for rpitx that makes RF transmission elegant

---

_Built with spite using https://github.com/psyb0t/servicepack_

---

_Now get out there and start broadcasting like the RF pirate you were meant to be! 🏴‍☠️📡_
