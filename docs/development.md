---
title: "Development"
description: "Working on PIrateRF: the code layout, the Make targets for local development and Pi management, tests and coverage, and building this site."
nav_order: 11
---

# Development

PIrateRF is built on [servicepack](https://github.com/psyb0t/servicepack). The framework owns `Makefile.servicepack`, the `Dockerfile.servicepack*` files and `scripts/make/servicepack/`; project targets live in [`Makefile`](https://github.com/psyb0t/piraterf/blob/main/Makefile).

## Layout

```text
cmd/                                 entry point
internal/pkg/services/piraterf/      the service
  http_server.go                     routes and the index page
  websocket*.go                      WebSocket events (rpitx, files, audio, presets, live mic)
  execution_manager.go               one transmission at a time, output streaming
  audio_processing.go                ffmpeg conversion to 16-bit 48 kHz mono WAV
  image_processing.go                image conversion for SSTV and Spectrum Paint
  file_upload_postprocessor.go       per-module upload handling
scripts/                             Pi setup, deploy and run scripts
html/                                index.html
static/                              CSS, JavaScript, fonts and images
files/                               runtime files: audio, images, data, IQ, presets
uploads/                             upload staging
docs/                                this site
```

## Make targets

| Target | Does |
| --- | --- |
| `make run-dev` | Runs PIrateRF in the development container on `127.0.0.1:8080` and `127.0.0.1:8443`, without rpitx. |
| `make build` | Cross-compiles the binary for the Pi Zero (ARM) into `build/`. |
| `make lint`, `make lint-fix` | Lints, or lints and fixes, the Go code. |
| `make test`, `make test-coverage` | Runs the tests; the coverage target fails below `MIN_TEST_COVERAGE` (70). |
| `make tls` | Generates the TLS certificate. |
| `make pi` | Full automated Pi setup. See [Installation from source](installation.md). |
| `make deploy`, `make install`, `make uninstall` | Copies the build to the Pi, installs or removes the systemd service. |
| `make ssh`, `make pi-reboot` | Opens a shell on the Pi, or reboots it. |
| `make pi-image` | Clones and shrinks the Pi's SD card into an image. |
| `make docs` | Builds this site into `.testing/docs-site`. |
| `make help` | Every target, including the servicepack ones. |

## Dependencies

- [rpitx](https://github.com/F5OEO/rpitx), through the [psyb0t/rpitx](https://github.com/psyb0t/rpitx) fork that `make pi` installs, generates the RF.
- [gorpitx](https://github.com/psyb0t/gorpitx) wraps rpitx's programs as Go modules.
- [aichteeteapee](https://github.com/psyb0t/aichteeteapee) provides the HTTP server, middleware and WebSocket hub.
