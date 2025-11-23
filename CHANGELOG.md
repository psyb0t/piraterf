# CHANGELOG

All notable changes to this pirate-ass project will be documented in this fuckin file.

## 2025-11-23

### Added - IQ Transmission Mode 📡

Finally added the IQ module - now you can replay raw IQ sample files and reproduce any RF signal you've captured. Perfect for repeater captures, signal analysis, and general RF fuckery.

**New Features:**

- **IQ File Upload & Management**: Upload .iq files through the web interface with rename/delete support
- **Full Parameter Control**:
  - Sample rate: 10 kHz to 2 MHz (auto-decimation above 200 kHz)
  - IQ data types: u8, i16 (default), float, double
  - Harmonic selection for frequency multiplication
  - Power level control (0.0 - 7.0)
  - Shared memory token support for runtime control
  - Configurable timeout (default 30 seconds)
  - Loop mode for continuous replay
- **File Management**: Dedicated `files/iqs/uploads/` directory with .iq extension filtering
- **State Persistence**: All settings saved to localStorage like a proper fuckin web app
- **Real-time Feedback**: WebSocket integration for live transmission status

**Backend:**

- Added `iqFilePostprocessor` for proper file handling
- IQ directory structure creation in service initialization
- Frontend env.js config generation for IQ paths
- Full integration with gorpitx SENDIQ module

**Frontend:**

- New IQ module form with all configuration options
- File upload, rename, and delete functionality
- Inline control layout matching other modules
- Validation ensuring required fields are set
- Help text explaining shared memory token behavior

**Updated Documentation:**

- README now lists 12 transmission modes (was 11)
- Added complete IQ section with configuration details
- Removed "Add RAW SendIQ module" from TODO (because we just fuckin did it)

This shit is ready to replay whatever RF signals you've captured. Record something with your SDR, save it as a .iq file, upload it to PIrateRF, and broadcast that fucker right back out. Signal reproduction has never been easier for the DIY RF pirate! 🏴‍☠️📡
