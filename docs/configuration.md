---
title: "Configuration"
description: "PIrateRF environment variables for directories, the HTTP and HTTPS listeners, logging and rpitx, plus the Pi and WiFi access point settings used at setup."
nav_order: 8
---

# Configuration

PIrateRF reads its settings from environment variables. On the Pi, [`scripts/piraterf.sh`](https://github.com/psyb0t/piraterf/blob/main/scripts/piraterf.sh) sets them before it starts the binary, so that script is where you change them.

## Contents

- [Directories](#directories)
- [HTTP and HTTPS](#http-and-https)
- [Logging and runtime](#logging-and-runtime)
- [Pi and access point settings](#pi-and-access-point-settings)

## Directories

Defined in [`internal/pkg/services/piraterf/config.go`](https://github.com/psyb0t/piraterf/blob/main/internal/pkg/services/piraterf/config.go).

| Variable | Default | What it holds |
| --- | --- | --- |
| `PIRATERF_HTMLDIR` | `./html` | `index.html` for the web UI |
| `PIRATERF_STATICDIR` | `./static` | CSS, JavaScript, fonts and images, served at `/static` |
| `PIRATERF_FILESDIR` | `./files` | Audio, images, data, IQ files and presets, served at `/files` |
| `PIRATERF_UPLOADDIR` | `./uploads` | Staging directory for uploads |

## HTTP and HTTPS

The web server comes from [aichteeteapee](https://github.com/psyb0t/aichteeteapee).

| Variable | Default | Set by `piraterf.sh` |
| --- | --- | --- |
| `HTTP_SERVER_LISTENADDRESS` | `127.0.0.1:8080` | `0.0.0.0:80` |
| `HTTP_SERVER_TLSENABLED` | `false` | `true` |
| `HTTP_SERVER_TLSLISTENADDRESS` | `127.0.0.1:8443` | `0.0.0.0:443` |
| `HTTP_SERVER_TLSCERTFILE` | empty | `./.tls/cert.pem` |
| `HTTP_SERVER_TLSKEYFILE` | empty | `./.tls/key.pem` |
| `HTTP_SERVER_READTIMEOUT` | `15s` | |
| `HTTP_SERVER_READHEADERTIMEOUT` | `10s` | |
| `HTTP_SERVER_WRITETIMEOUT` | `30s` | |
| `HTTP_SERVER_IDLETIMEOUT` | `60s` | |
| `HTTP_SERVER_MAXHEADERBYTES` | `1048576` | |
| `HTTP_SERVER_SHUTDOWNTIMEOUT` | `10s` | |
| `HTTP_SERVER_SERVICENAME` | `http-server` | |
| `HTTP_SERVER_FILEUPLOADMAXMEMORY` | `33554432` (32 MB) | |

`piraterf.sh` generates a self-signed certificate in `./.tls/` the first time it runs, which is why the browser warns about it.

## Logging and runtime

| Variable | Default | Set by `piraterf.sh` |
| --- | --- | --- |
| `LOG_LEVEL` | `info` | `warn` |
| `LOG_FORMAT` | `text` | `json` |
| `LOG_CALLER` | `false` | `true` |
| `ENV` | | `prod` |
| `GORPITX_PATH` | `$HOME/rpitx` | `/home/<PI_USER>/rpitx` |
| `SERVICES_ENABLED` | all services | |

`GORPITX_PATH` is the rpitx checkout that [gorpitx](https://github.com/psyb0t/gorpitx) runs. `SERVICES_ENABLED` is the servicepack switch that limits which services start.

## Pi and access point settings

[`scripts/pi_config.sh`](https://github.com/psyb0t/piraterf/blob/main/scripts/pi_config.sh) holds the values the setup and deploy scripts use. Edit it before `make pi`; see [Installation from source](installation.md).

| Variable | Default | Used for |
| --- | --- | --- |
| `PI_USER` | `fucker` | SSH user on the Pi |
| `PI_HOST` | `piraterf.local` | Hostname or address of the Pi |
| `PI_PASS` | `FUCKER` | SSH password |
| `AP_SSID` | `🏴‍☠️📡` | WiFi network name |
| `AP_PASSWORD` | `FUCKER!!!` | WiFi password |
| `AP_CHANNEL` | `7` | WiFi channel |
| `AP_COUNTRY` | `US` | WiFi regulatory country |

These defaults are public. Change the passwords; [Security](security.md) explains why.
