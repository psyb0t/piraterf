---
title: "Home"
description: "PIrateRF turns a Raspberry Pi Zero W into a portable RF transmitter with its own WiFi hotspot and a browser UI for 12 transmission modes."
permalink: /
nav_order: 1
seo:
  type: WebSite
---

# PIrateRF - Software-Defined Radio Transmission Platform

PIrateRF turns a **Raspberry Pi Zero W** into a portable RF signal generator that spawns its own WiFi hotspot. You connect a phone or laptop to that hotspot, open the web UI and transmit anything from FM broadcasts with RDS to FT8, POCSAG pager messages, SSTV pictures and raw IQ replays.

Under the hood it is one Go service that drives [rpitx](https://github.com/F5OEO/rpitx) through the [gorpitx](https://github.com/psyb0t/gorpitx) wrapper. rpitx generates the RF signal straight from **GPIO 4 (physical pin 7)**, so the only hardware you need is the Pi itself and, for anything past your own room, a filter and an antenna.

{: .warning }
PIrateRF is for amateur radio experimentation and education, including safe indoor testing without an external antenna. You are responsible for following your local RF regulations and licensing rules. Read [Legal and safety](legal-and-safety.md) before you put an antenna on it.

## Fastest way in

1. Download the pre-built SD card image and flash it.
2. Boot the Pi Zero W and join the WiFi network `🏴‍☠️📡` (password `FUCKER!!!`).
3. Open `https://piraterf.local` and accept the self-signed certificate.

[Getting started](getting-started.md) has every step, including the download link.

## You want to, read

| You want to | Read |
| --- | --- |
| Get a working transmitter in ten minutes | [Getting started](getting-started.md) |
| Build and deploy from source to your own Pi | [Installation from source](installation.md) |
| Know what each of the 12 modes does and how to receive it | [Transmission modes](transmission-modes.md) |
| Save and reload mode settings | [Presets](presets.md) |
| Pick a frequency | [Standard frequencies](frequencies.md) |
| Hook up an antenna without spraying harmonics everywhere | [Legal and safety](legal-and-safety.md) |
| Change ports, directories, logging or the WiFi network | [Configuration](configuration.md) |
| Drive PIrateRF from your own client | [HTTP and WebSocket API](api.md) |
| Know what the box exposes before you take it somewhere | [Security](security.md) |
| Change the code or run the tests | [Development](development.md) |

## Videos

- [PIrateRF promo video](https://www.youtube.com/watch?v=hnYkvDLOi70)
- [PIrateRF showcase video](https://www.youtube.com/watch?v=7DeMhe47aNQ)
