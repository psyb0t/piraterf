---
title: "Legal and safety"
description: "Antenna options for PIrateRF from no antenna to filtered amplifiers, why the Pi's square wave needs a low-pass filter, and the licensing rules you must follow."
nav_order: 7
---

# Legal and safety

PIrateRF is built for amateur radio experimentation and education, including safe indoor testing without an external antenna. You are responsible for following your local RF regulations and licensing rules. All demonstration images in this documentation were made indoors without an antenna, at a range of about 5 meters.

## Antenna setup

The RF output is **GPIO 4 (physical pin 7)** on the Pi Zero W.

| Setup | What you get |
| --- | --- |
| No antenna | An extremely weak signal that stays inside your home. Safe for testing. |
| Short wire (10 to 20 cm) | Minimal range for indoor experiments. |
| Wire antenna (75 cm) | Longer range, but the square wave's harmonics travel farther too. Keep it indoors. |
| Low-pass filter + antenna | Proper outdoor transmission, with a license. |
| Low-pass filter + amplifier + antenna | Maximum range. The Pi itself outputs only milliwatts. |

{: .warning }
The GPIO pin outputs a square wave with harmonics across the whole spectrum. Without a low-pass filter for your operating frequency you transmit on many frequencies at once. Outdoor transmission needs that filter.

## Licensing

- Most frequencies require an amateur radio license. FT8, RTTY, FSK, SSTV and most HF, VHF and UHF operation need one. Check with your national authority (FCC, Ofcom and so on).
- Stay inside the amateur bands. [Standard frequencies](frequencies.md) lists the common allocations, and band plans differ between countries.
- Amateur radio prohibits commercial content.
- Never interfere with emergency or public safety communications.

## Hardware for proper use

- **Low-pass filter:** required for anything beyond indoor testing.
- **Antenna:** use one that is resonant on your operating frequency.
- **Power:** the Pi outputs a few milliwatts. With an amplifier, respect the power limit of your license class.
- **SWR:** at milliwatt levels the Pi does not care. An amplifier does, so match the impedance or you damage it.

## Indoor testing

Without an antenna the signal stays inside your home. That is enough to learn RF concepts and digital modes, and to test encoding and decoding with a receiver in the same room.
