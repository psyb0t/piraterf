---
title: "Getting started"
description: "Flash the pre-built PIrateRF image to a Raspberry Pi Zero W, join its WiFi hotspot and start transmitting from your browser."
nav_order: 2
---

# Getting started

The pre-built image is the shortest path from nothing to a working transmitter. It is a full Raspberry Pi OS install with rpitx, the WiFi access point and the PIrateRF service already set up. If you would rather build everything yourself, read [Installation from source](installation.md) instead.

## What you need

- A **Raspberry Pi Zero 1 W** (the original model) and an SD card of 4 GB or more.
- A computer with an SD card reader.
- A phone, tablet or laptop with WiFi and a browser to control it.

rpitx itself runs on the Pi 1, 2, 3, 3B+, Zero, Zero W and Pi 4, but the deployment scripts, the access point setup, the ARM build and the pre-built image all target the Pi Zero 1 W. Other boards may work with manual setup. I have not tested them.

## Steps

1. Download the image: [PIrateRF image v2025-11-23](https://archive.org/download/piraterf-2025-11-23-22-18-49-20251123202604/piraterf_2025-11-23_22-18-49.img).
2. Flash it to the SD card. With **Raspberry Pi Imager**, pick "Use custom image" and choose the downloaded `.img` file. On the command line, replace `/dev/sdX` with your SD card device:

   ```bash
   sudo dd \
     if=piraterf_2025-11-23_22-18-49.img \
     of=/dev/sdX \
     bs=4M \
     status=progress
   ```

3. Put the card in the Pi Zero W and power it on. Give it a minute to boot.
4. Join the WiFi network `🏴‍☠️📡` with the password `FUCKER!!!`.
5. Open `https://piraterf.local` in your browser. The certificate is self-signed, so the browser warns you once; accept it to continue.
6. Pick a mode from the dropdown, fill in the frequency and the mode's settings, and press start.

## Default settings on the image

| What | Value |
| --- | --- |
| WiFi access point | SSID `🏴‍☠️📡`, password `FUCKER!!!` |
| SSH | user `fucker`, password `FUCKER`, host `piraterf.local` |
| Web UI | `https://piraterf.local` (plain HTTP on port 80 works too) |

These are public defaults. Change at least the WiFi password before you take the Pi anywhere other people are; [Security](security.md) explains what an attacker on that network can do.

## Sharing one transmitter

Any number of devices can join the access point and open the UI at the same time. Only one transmission runs at a time, because there is one GPIO pin to drive. Every connected device sees the transmission start and stop, sees the live output of the rpitx process, and can stop a transmission someone else started or start the next one.

## Next steps

- [Transmission modes](transmission-modes.md) explains every mode and how to receive it with an SDR.
- [Legal and safety](legal-and-safety.md) covers antennas, filters and the rules you are expected to follow.
- [Presets](presets.md) saves mode settings so you do not type them again.
