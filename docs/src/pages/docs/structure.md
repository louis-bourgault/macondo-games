---
layout: '../../layouts/DocsLayout.astro'
title: 'Project Structure'
draft: false
category: "general"
---

## File structure

## Project Structure
- /software - all software for the Go game system and the current embedded MicroPython firmware
- /software/cmd/embed - the current Go-hosts-MicroPython firmware and its build system
- /mini-games - The kicad project and files for the version one hardware
- /mini-games/jlcpcb/production_files - The files for sending version one to JLC, exported by the jlcpcb tools plugin in kicad. Big thanks to this plugin: [Bouni/kicad-jlcpcb-tools](https://github.com/Bouni/kicad-jlcpcb-tools)
- /hardware-v2 - The kicad project and files for the version two RP2350 hardware, which is still in development. Its JLC files are under /hardware-v2/jlcpcb/production_files.
- /case - files for making the cases. Version one and version two are in their own subdirectories.
- /img - the images used in the main readme
- /docs - this website
- /mp-sdk - the legacy MicroPython SDK and its custom MicroPython board definition
- /mp-embedding - an early experiment used while working out how to embed MicroPython into Go
- /web-editor - the web editor for the current embedded MicroPython SDK.
