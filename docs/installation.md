---
title: "Installation from source"
description: "Build PIrateRF for the Pi Zero W, set up rpitx and the WiFi access point, deploy the systemd service, update, uninstall and clone the SD card."
nav_order: 3
---

# Installation from source

This route builds PIrateRF on your computer, cross-compiles it for the Pi Zero's ARMv6 CPU and pushes it to the Pi over SSH. Use it when you want your own WiFi name and password, your own user, or a build of your own code. If you just want a working transmitter, the [pre-built image](getting-started.md) is faster.

## Requirements

On your computer:

- Linux with `bash`, `make`, `docker`, `openssl`, `ssh`, `scp` and `sshpass`. The build runs the Go compiler inside the `golang:1.24.6-alpine` image, so you do not need Go installed. The deploy scripts log in with `sshpass` and the password from `scripts/pi_config.sh`.
- A clone of the repository.

On the Pi:

- A **Raspberry Pi Zero 1 W** running **Raspberry Pi OS Bullseye (Debian 11)**. The access point script checks `/etc/os-release` and refuses to run on any other version.
- SSH access and an internet connection, both covered below.

## Connect the Pi Zero first

**Before you do ANYTHING else**, get your Pi Zero W connected and reachable over SSH. This tutorial gets it working over USB:

[Pi Zero W USB connection tutorial](https://ciprian.51k.eu/pi-zero-1-wh-ubuntu-24-04-usb-connection-the-tutorial-that-actually-fkin-works/)

### Share your internet connection with the Pi

The setup scripts download packages, so the Pi needs internet through your computer. On Ubuntu or another Linux with NetworkManager:

1. In the network settings, open the USB connection (usually `usb0`), go to the IPv4 settings and change the method from "Link-Local Only" to "Shared to other computers".
2. Stop Docker, which interferes with the forwarding rules:

   ```bash
   sudo systemctl stop docker.socket
   sudo systemctl stop docker
   ```

3. Restart NetworkManager:

   ```bash
   sudo systemctl restart NetworkManager
   ```

4. Add forwarding and NAT rules. Replace `usb0` with the Pi's interface and `enp5s0` with your internet interface; `ip link show` lists both:

   ```bash
   sudo iptables -A FORWARD -i usb0 -o enp5s0 -j ACCEPT
   sudo iptables -A FORWARD -i enp5s0 -o usb0 -j ACCEPT
   sudo iptables -t nat -A POSTROUTING \
     -s 10.42.0.0/24 \
     -o enp5s0 \
     -j MASQUERADE
   ```

5. Test it: SSH into the Pi (`ssh fucker@piraterf.local`, or whatever user and host you set up) and run `ping 8.8.8.8`.

Start Docker again before you build, because `make build` needs it.

## Configure the Pi settings

Every script reads [`scripts/pi_config.sh`](https://github.com/psyb0t/piraterf/blob/main/scripts/pi_config.sh). Edit it before running anything:

```bash
git clone https://github.com/psyb0t/piraterf.git
cd piraterf
nano scripts/pi_config.sh
```

```bash
export PI_USER="fucker"             # Pi username
export PI_HOST="piraterf.local"     # Pi hostname or IP
export PI_PASS="FUCKER"             # Pi password, used by sshpass

# WiFi AP Configuration
export AP_SSID="🏴‍☠️📡"            # WiFi AP name
export AP_PASSWORD="FUCKER!!!"      # WiFi AP password (WPA2)
export AP_CHANNEL="7"               # WiFi channel
export AP_COUNTRY="US"              # WiFi regulatory country code
```

[Configuration](configuration.md#pi-and-access-point-settings) describes each value.

## Run the full setup

```bash
make pi
```

`make pi` runs these targets in order and stops at the first failure:

| Step | Target | What it does on the Pi |
| --- | --- | --- |
| 1 | `make pi-setup-deps` | Installs `git`, `sox`, `libsox-fmt-all`, `ffmpeg`, `imagemagick`, `openssl`, `minimodem`, `pulseaudio` and `socat`, then clones [psyb0t/rpitx](https://github.com/psyb0t/rpitx) to `/home/$PI_USER/rpitx` and runs its `install.sh`. |
| 2 | `make pi-setup-ap` | Runs `apt-get upgrade`, installs `hostapd` and `dnsmasq`, and turns `wlan0` into a WPA2 access point at `192.168.4.1` that hands out `192.168.4.2` to `192.168.4.50`. |
| 3 | `make pi-setup-branding` | Replaces `/etc/motd`, `/etc/bash.bashrc` and the user's `~/.bashrc` with the PIrateRF banner and prompt. |
| 4 | `make deploy` | Runs `make build` locally, copies `build/piraterf.tar.gz` to the Pi and unpacks it into `/home/$PI_USER/piraterf`. |
| 5 | `make install` | Writes and starts the `piraterf` systemd service. |
| 6 | `make pi-reboot` | Reboots the Pi into access point mode. |

Two steps ask a question when they find an earlier install. `pi-setup-deps` asks whether to reinstall when `/home/$PI_USER/rpitx` exists, and `deploy` asks whether to wipe `/home/$PI_USER/piraterf` or just overwrite the files in it. Answer them in the terminal.

After the reboot, join your access point, open `https://piraterf.local` and start transmitting.

### What the build produces

`make build` checks for `.tls/cert.pem` and `.tls/key.pem` and runs `make tls` to create a self-signed pair when they are missing. It then compiles a static `GOOS=linux GOARCH=arm GOARM=6` binary in Docker and packs `build/piraterf.tar.gz` with:

- the `piraterf` binary,
- `html/`, `static/` and the bundled sound effects in `files/audio/sfx/`,
- `.tls/` with the certificate and key,
- `piraterf.sh`, `install.sh`, `uninstall.sh` and `pi_config.sh` from `scripts/`.

The binary itself is removed afterwards; only the archive stays in `build/`.

## What lives where on the Pi

| Path | Contents |
| --- | --- |
| `/home/$PI_USER/piraterf/` | The binary, web UI, scripts and certificate. This is the service's working directory. |
| `/home/$PI_USER/piraterf/files/` | Uploaded audio, images, data and IQ files, and saved presets. |
| `/home/$PI_USER/piraterf/uploads/` | Upload staging, plus the unix sockets for live microphone broadcasts. |
| `/home/$PI_USER/rpitx/` | rpitx, which gorpitx calls for every transmission. |
| `/etc/systemd/system/piraterf.service` | The service unit. It runs `piraterf.sh` as root and restarts it 5 seconds after any exit. |
| `/home/$PI_USER/ap-backup-*/` | Copies of `dhcpcd.conf`, `dnsmasq.conf` and `hostapd.conf` from before the access point setup. |

The branding step backs up the files it replaces under `/tmp/piraterf_backups_<timestamp>/`, which does not survive the reboot at the end of `make pi`. Copy that directory somewhere else first if you want to keep the originals.

`files/` is where your work lives. Back it up before you run `make uninstall` or answer yes to the wipe question in `make deploy`.

## Updating

Pull or change the code and run:

```bash
make deploy
```

`make deploy` rebuilds the archive, unpacks it over the existing install and restarts the service when it is enabled. Answer `N` to the overwrite question to keep `files/`.

## Single steps

```bash
make pi-setup-deps       # rpitx and system packages
make pi-setup-ap         # WiFi access point
make pi-setup-branding   # banner and prompt
make deploy              # build and copy to the Pi
make install             # install and start the systemd service
make pi-reboot           # reboot the Pi
make ssh                 # SSH into the Pi with the configured password
make uninstall           # stop the service and delete /home/$PI_USER/piraterf
```

`make install` refuses to run when `/etc/systemd/system/piraterf.service` already exists. Run `make uninstall` first. `make uninstall` leaves rpitx, the access point and the branding in place.

## Cloning your SD card into an image

```bash
make pi-image
```

`make pi-image` scans `/dev/sd*` and `/dev/mmcblk*` for disks with a Raspberry Pi boot and root partition, asks which one holds your Pi's SD card, copies it with `dd` to `images/piraterf_<date>_<time>.img` and shrinks the result with [PiShrink](https://github.com/Drewsif/PiShrink). It needs `sudo`, and asks for the path to `pishrink.sh` when it is not on your `PATH`. Flash the image to other cards with Raspberry Pi Imager or `dd`.
