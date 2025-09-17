# 🏴‍☠️ PIrateRF - Software-Defined Radio Transmission Platform

**PIrateRF** is a fucking badass software-defined radio (SDR) transmission platform that turns your **Raspberry Pi Zero W** into a portable RF signal generator with a sleek web interface. This beast enables you to transmit various types of radio signals including FM radio broadcasts, Morse code, carrier waves, and even spectrum painting - all controlled through your browser like a proper pirate! 📡⚡

## 🎯 What the Fuck Does This Thing Do?

PIrateRF transforms your Pi Zero into a **standalone RF transmission station** that can:

- **🎵 FM Radio Broadcasting**: Transmit audio with full RDS (Radio Data System) metadata including station names, radio text, and PI codes
- **📻 Morse Code Transmission**: Send CW (continuous wave) Morse code signals
- **🎛️ Carrier Wave Generation**: Simple tone generation for testing and tuning
- **🎨 Spectrum Painting**: Transmit images as RF spectrum patterns (because why the fuck not?)
- **🎧 Real-time Audio Processing**: Upload files or record directly through the browser
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

### 1. Initial Pi Setup and Configuration

Flash Raspberry Pi OS Lite to your SD card and enable SSH. Then:

```bash
# Clone this badass project
git clone https://github.com/psyb0t/piraterf.git
cd piraterf

# Edit scripts/pi_config.sh and modify these values to match your Pi:
# export PI_USER="fucker"              # Pi username
# export PI_HOST="piraterf.local"      # Pi hostname/IP
# export PI_PASS="FUCKER"             # Pi password
```

### 2. Complete Automated Setup

Run the full setup pipeline that configures everything automatically:

```bash
make complete
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
2. **Open browser**: Navigate to `https://piraterf.local` **ONLY** (don't use the IP address!)
3. **Start transmitting**: Upload audio, configure RDS, and broadcast like a proper pirate!

**⚠️ IMPORTANT**: Use `https://piraterf.local` **NOT** the IP address (`192.168.4.1`). The fucking microphone recording feature requires HTTPS with a proper hostname to work due to browser security restrictions. Using the IP address will break microphone access!

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

### FM Radio Broadcasting (PIFMRDS)

- **Audio Support**: MP3, WAV, FLAC, and more (auto-converted to 48kHz/16-bit/mono)
- **RDS Metadata**: Station name, radio text, PI codes, and program type
- **Playlist Mode**: Create playlists with intro/outro support
- **Play Once Mode**: Single play with automatic timeout calculation
- **Live Recording**: Record audio directly through the browser interface

### Morse Code (CW)

- **Text to Morse**: Automatic conversion with configurable WPM (words per minute)
- **Custom Messages**: Send any text as Morse code
- **Frequency Control**: Adjustable carrier frequency for different bands

### Tune Mode

- **Carrier Wave**: Simple tone generation for testing and frequency verification
- **Frequency Sweep**: Testing and calibration support

### Spectrum Painting

- **Image Upload**: Convert images to RF spectrum patterns
- **Format Support**: JPEG, PNG, GIF with automatic YUV conversion
- **Visual RF**: Turn your images into radio art (because pirates love art too)

## 🌐 Network Configuration

The Pi automatically configures itself as a **standalone WiFi access point**:

- **SSID**: "🏴‍☠️📡"
- **Password**: "FUCKER!!!"
- **IP Range**: 192.168.4.1/24 (Pi is at 192.168.4.1)
- **DHCP**: Automatic IP assignment for connected devices
- **Web Interface**:
  - **Primary**: `https://piraterf.local` (port 443) - **USE THIS ONE**
  - **Fallback HTTP**: `http://192.168.4.1` (port 80) - limited functionality
  - **Fallback HTTPS**: `https://192.168.4.1` (port 443) - microphone won't work

## 🔒 Security Features

- **Self-signed TLS**: Auto-generated certificates for HTTPS
- **Isolated Network**: Pi runs its own WiFi network
- **File Upload Security**: Secure multipart upload with validation
- **No Authentication**: Designed for standalone/isolated use (add your own if needed)

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
│   └── make/              # Build and deployment scripts
│       ├── build.sh       # Cross-compilation for Pi
│       ├── pi_setup_*.sh  # Pi configuration scripts
│       ├── deploy.sh      # Deployment automation
│       └── servicepack/   # Framework scripts
├── html/                  # Web interface templates
├── static/                # CSS, JavaScript, images
├── files/                 # Audio and image file storage
├── uploads/               # Temporary upload staging
├── .tls/                  # TLS certificates
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
- `make complete` - Run full setup pipeline
- `make ssh` - SSH into the Pi
- `make pi-reboot` - Reboot the Pi
- `make uninstall` - Remove PIrateRF from Pi

### Utilities

- `make tls` - Generate TLS certificates
- `make clean` - Clean build artifacts
- `make help` - Show all available targets

## 🔧 Configuration

### Pi Connection Settings

The Pi connection settings are defined in `scripts/pi_config.sh` (configured in setup step 1).

### Service Configuration

The PIrateRF service uses environment variables for configuration:

```bash
PIRATERF_HTMLDIR=/path/to/html       # Web templates directory
PIRATERF_STATICDIR=/path/to/static   # Static assets directory
PIRATERF_FILESDIR=/path/to/files     # Audio/image file storage
PIRATERF_UPLOADDIR=/path/to/uploads  # Upload staging directory
```

## 🎵 Audio Processing Pipeline

PIrateRF automatically processes uploaded audio through a sophisticated pipeline:

1. **Format Detection**: Supports MP3, WAV, FLAC, OGG, and more
2. **Conversion**: Automatically converts to 48kHz, 16-bit, mono WAV using Sox
3. **Validation**: Ensures audio meets RF transmission requirements
4. **Storage**: Organizes files in `/files/audio/uploads/` and `/files/audio/sfx/`
5. **Playlist Support**: Create playlists with intro/outro and repeat modes

## 🖼️ Image Processing for Spectrum Painting

Images are processed for RF spectrum transmission:

1. **Format Support**: JPEG, PNG, GIF automatically detected
2. **Conversion**: Converted to YUV format for RF transmission
3. **Optimization**: Resized and optimized for spectrum display
4. **Storage**: Organized in `/files/images/uploads/`

## 🏴‍☠️ Legal and Safety Notice

**IMPORTANT**: This software enables RF transmission. **You are responsible for complying with your local RF regulations and licensing requirements.**

- Ensure you have proper licenses for your transmission frequency and power levels
- Some frequencies require amateur radio licenses
- Respect power limitations and spurious emission requirements
- Don't interfere with emergency services or licensed operators
- When in doubt, consult your local RF regulatory authority

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

This project is licensed under the terms specified in the `LICENSE` file.

## 🙏 Acknowledgments

- **[rpitx](https://github.com/F5OEO/rpitx)** - The legendary RF transmission library that makes this all possible
- **[servicepack](https://github.com/psyb0t/servicepack)** - The framework that keeps this project organized and deployable
- **[aichteeteapee](https://github.com/psyb0t/aichteeteapee)** - The HTTP server framework powering the web interface
- **[gorpitx](https://github.com/psyb0t/gorpitx)** - Go wrapper for rpitx that makes RF transmission elegant
- **Go community** - For building such a fucking excellent language
- **Raspberry Pi Foundation** - For creating the perfect pirate ship hardware

---

_Built with spite using https://github.com/psyb0t/servicepack_

---

_Now get out there and start broadcasting like the RF pirate you were meant to be! 🏴‍☠️📡_
