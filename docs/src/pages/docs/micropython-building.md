---
layout: '../../layouts/DocsLayout.astro'
title: 'Building the current MicroPython SDK'
draft: false
category: "mp"
---

The current MicroPython firmware is built from ```/software/cmd/embed```. This is the version used by the web editor, not the older firmware under ```/mp-sdk```.

## Setup

You will need:

- Go
- TinyGo
- just
- make
- gcc and binutils for a host build
- the Arm GNU toolchain (```arm-none-eabi-gcc``` and ```arm-none-eabi-ar```) for a device build
- a full clone of MicroPython

On macOS, most of these can be installed with Homebrew. The package names may be different on Linux.

```sh
xcode-select --install
brew install go tinygo just
brew install --cask gcc-arm-embedded
```

Clone MicroPython somewhere outside this repository:

```sh
git clone https://github.com/micropython/micropython.git ~/src/micropython
```

The build uses ```~/src/micropython``` by default. If you put it somewhere else, set ```MICROPYTHON_ROOT``` when running the commands below.

## Building for the device

From the embedded firmware directory, run:

```sh
cd software/cmd/embed
just device
```

If your MicroPython checkout is somewhere else:

```sh
cd software/cmd/embed
MICROPYTHON_ROOT=/path/to/micropython just device
```

This produces ```software/cmd/embed/ferret.uf2```. The default target is version one, using the RP2040 and 16MB flash configuration in ```ferretboard.json```.

To build the version two RP2350 target instead, use:

```sh
just --set TARGET ferretboard2.json device
```

You can copy the UF2 onto the board while it is in BOOTSEL mode, or flash it with picotool:

```sh
picotool load ferret.uf2 -f
```

## Building the host reference

There is also a host build with a MicroPython REPL over your normal terminal. Drawing is stubbed out, but it is useful for checking that the interpreter and ```ferret``` module build correctly.

```sh
cd software/cmd/embed
just host
./spike
```

The host build creates ```spike``` in the same directory. Its fake flash filesystem is stored in ```build/ferret.img``` by default.

## Cleaning

To remove the generated archives and executables:

```sh
cd software/cmd/embed
just clean
```

For a much more detailed explanation of how all the C, Go and MicroPython parts are compiled and linked together, see ```/software/cmd/embed/BUILDING.md``` in the repository.
