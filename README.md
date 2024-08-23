# nv-reloaded
An open-source rework in Go of my dead-born n0vel1st project

## Description

(TODO)

## Status

- **2024-05-05**: Development is currently in alpha state. This code
does NOT allow a practical use of the device. If you're
interested in participating in code development, now is a good
time. If you're more interested in building your own device and
using it, you likely will have to wait for a few more months.

## Discord Server

If you want to participate in discussions about this project,
join the "NV Reloaded" discord server at [Discord][discord].
Snapshots, videos, and regular status updates are posted there.

## Stable Releases

Stable releases will be announced here:
- ...
- 2024-05-05: no current stable release.

## Installing (prototyping and development)

**These steps are for tests and development**. A simpler
process will be defined later on.

<hr>

### NOTES

This prototype runs on an 8GB Raspberry Pi4 with
Raspberry Pi OS 64-bits, AKA Debian 12, AKA "bookworm" (nice coincidence)

It may run on a Pi4 with less memory (possibly down to
2GB), but that's not tested. It may also run on a Pi3 or
inferior, but again, that's not tested.

It has also not been tested on the new Pi5. The extra power
of a Pi5 may not make a big difference, since the only
bottleneck currently is in the communication with the
e-ink display through the e-Ink Hat, which is limited
to SPI running at 12.5MHz max.

<hr>

### Connect accessories

* e-Ink Hat
* e-Ink display
* keyboard if not BT/BLE

### Install Pi and create a user account on your Pi4

Your user should be created during the Pi's Install
Procedure (check [here][pi-install] for more details).
Also ensure you set up your wifi properly (it will be
possible to set wifi later on with the display, but
this is not yet done). Also ensure you can access your
pi through SSH if you're not going to use an external
monitor.

### Set up Pi config

Using `sudo raspi-config`, ensure the SPI port is ON

### Install necessary tools

```
sudo apt update
sudo apt install -y git build-essential evtest libfreetype6 libfreetype2-dev libfreetype6-dev
(possibly more... TBD)
```

### Install NV-Reloaded code base

Download the main code base from Github into your home
directory on the pi:
```
cd ~
git clone https://github.com/peergum/nv-reloaded nv
cd nv
make
```

### Installing go 1.22.2 on Pi4 (64 bits)

```
cd ~
./install_go
```

### Run it

#### with default debug options
```
make run
```

#### Or however you want to run it
```
./nv -h
./nv -debug -display -epd -doc
./nv [...]
```

[pi-install]: https://www.raspberrypi.com/software/
[discord]: https://discord.gg/FJxdYGMF