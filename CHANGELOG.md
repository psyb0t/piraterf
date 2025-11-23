# CHANGELOG

All notable changes to this pirate-ass project will be documented in this fuckin file.

## 2025-11-23

### Added - Preset System 💾

Finally added a proper preset system so you don't have to keep entering the same fuckin parameters every time. Save your favorite configs, reload them with one click, and stop wasting time re-entering frequencies and settings.

**New Features:**

- **Save/Load Presets**: Save current module configuration as named presets, load them back with one click
- **Preset Management**:
  - Create new presets with custom names
  - Rename existing presets
  - Delete presets you don't need anymore
  - Reload preset values to revert unsaved changes
- **Per-Module Storage**: Each module has its own preset directory (`files/presets/{module}/`)
- **Persistent Selection**: Last selected preset is remembered across page refreshes (but doesn't auto-load to avoid fucking up your form)
- **Success Notifications**: Green popup notifications when presets are created, saved, or reloaded
- **File-Based Storage**: Presets stored as JSON files, easy to backup or share
- **Auto Directory Creation**: All preset directories created automatically at startup for all 12 supported modules

**Backend:**

- WebSocket handlers for preset operations: `preset.save`, `preset.load`, `preset.rename`, `preset.delete`
- JSON file storage in `files/presets/{modulename}/` structure
- Directory permissions fixed to `0o755` for proper web serving
- File permissions set to `0o644` for readability

**Frontend:**

- Preset dropdown selector with refresh button
- Save, Reload, and Edit buttons (enabled when preset is selected)
- New preset creation modal
- Edit preset modal with rename/delete options
- LocalStorage persistence for last selected preset
- Success/error notifications for all operations

**Fixed:**

- Timeout placeholders changed from "30" to "0" to reflect actual defaults (PIFMRDS and SENDIQ modules)
- Preset dropdown now correctly populates from JSON file listing API
- Button states properly managed across page refreshes

Now you can save your favorite doorbell replay configs, your go-to FM broadcast settings, or that perfect FT8 setup, and load them back in a fuckin second. No more re-entering the same shit over and over! 🏴‍☠️💾

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
