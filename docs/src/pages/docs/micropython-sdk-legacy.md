---
layout: '../../layouts/DocsLayout.astro'
title: 'MicroPython SDK (legacy) - About'
draft: false
category: "ogmp"
---

Apart from the golang sdk, there is also a micropython sdk. This is designed for ease of use, even to go so far as to be an introduction to programming.
This came about after I showed this project to a digital technologies teacher at my school, and he was somewhat interested in how something like the device could be used in a school setting due to its simplicity and people being able to make something that they can actually hold in their hands in a relatively limited amount of time. However, coding in Go doesn't really work for this: it's not that beginner friendly a language (for someone who's never coded before, at least), and more importantly there would be a significant amount of software that would need to be installed on the computers for people to even start working on it, including VS code and the compiler.

So I made a simple micropython sdk that was programmed through an earlier version of the web editor.

Note: this is the old version of the mp-sdk. The current deployed editor uses the [new embedded SDK](/docs/micropython-sdk), so it should not be used with this firmware.

## Building the micropython sdk.
You will need an Arm GCC toolchain, cmake and ninja. On macOS, I installed them with:

```sh
brew install --cask gcc-arm-embedded
brew install cmake ninja
```

Then clone MicroPython and prepare its RP2 port:

```sh
git clone https://github.com/micropython/micropython.git
cd micropython
make -C mpy-cross
cd ports/rp2
make submodules
```

From ```ports/rp2```, build using the board and manifest from this repository. Replace ```/path/to/macondo-games``` with the real path to your clone:

```sh
make BOARD_DIR=/path/to/macondo-games/mp-sdk/boards/ferretboard \
  FROZEN_MANIFEST=/path/to/macondo-games/mp-sdk/boards/ferretboard/manifest.py
```

This creates the legacy UF2 in MicroPython's ```ports/rp2/build-ferretboard``` directory.
