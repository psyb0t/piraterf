---
title: "Security"
description: "What PIrateRF exposes on its WiFi network: no authentication, file rename and delete on any path as root, public default passwords and a self-signed certificate."
nav_order: 10
---

# Security

PIrateRF is a toy for a network you control. Anyone who can join its WiFi network can use everything it offers, and that is more than transmitting. Know this before you take it somewhere other people are.

## What anyone on the network can do

- **Use every route and event.** No HTTP route and no WebSocket event checks who is calling. Anyone on the network can start and stop transmissions, upload files and manage presets.
- **Rename or delete any file on the Pi.** `file.rename` and `file.delete` take a full filesystem path and do not restrict it to the files directory. The service runs as root (`User=root` in the systemd unit [`scripts/install.sh`](https://github.com/psyb0t/piraterf/blob/main/scripts/install.sh) installs), so this covers the whole system.
- **Write preset files outside the presets directory.** Preset module and preset names are joined into a path without checks, so a crafted name can point the `.json` file elsewhere.
- **Reach it from another site.** The CORS middleware allows any origin, so a web page open on a connected device can call the HTTP routes.

## Defaults to change

- The WiFi password, SSH user and SSH password in [`scripts/pi_config.sh`](https://github.com/psyb0t/piraterf/blob/main/scripts/pi_config.sh) and in the pre-built image are public. Change them; see [Configuration](configuration.md#pi-and-access-point-settings).
- HTTPS uses a self-signed certificate generated on the Pi, so browsers warn and cannot tell a real PIrateRF from an impostor. Plain HTTP stays open on port 80.

## Keeping it contained

Run it on its own access point, as `make pi` sets it up, and only let people you trust join that network. Do not bridge it to a network you care about.
