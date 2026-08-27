![banner](/img/banner.png)
# ferretboard!!!

> [!IMPORTANT]
> # PLEASE READ
> The primary documentation for this project can be found at [ferretboard.louisbourgault.com/docs](https://ferretboard.louisbourgault.com/docs). Please read from that source for organised information on software interfaces, assembly, and a more complete project overview.
> 
> This README remains the authoritative source on commands to build the firmware for the device, and also has a nice image gallery at the end.

[![View PCB on KiCanvas](https://hack.club/pcb-badge)](https://kicanvas.org/?repo=https://github.com/louis-bourgault/macondo-games/tree/main/hardware-v2)

# What is this?
Ferretboard is a simple games console, made to be easy to program and cheap to produce. Version one uses a RP2040 MCU; version two, currently in development, upgrades this to the RP2350 and changes the form factor significantly. The display is a 240x240 1.54" LCD, which is common and cost effective. It is battery powered, using a basic 3.7v LiPo.

The idea for this project, in the long term, is to be an easy introduction to programming and game development for students. The micropython sdk is editable in web, easy to get started using, and would be perfect in the long run for a classroom setting or Hack Club larger program. The upgrade to the RP2350 will make this level of software possible, with more RAM and processing power to make running complicated games within Micropython possible. 

The overall dimensions of the build are around 100x50mm for both v1 and v2, with a depth of around 20mm. The case is 3d printed.

This is made as a part of Hack Club Macondo. You can find my project at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137). Thank you to Hack Club for financing and helping me with this project. You can find all acknowledgements for this project at [ferretboard.louisbourgault.com/thanks](https://ferretboard.louisbourgault.com/thanks)

## Online stuff!
Docs site and main site: [ferretboard.louisbourgault.com](https://ferretboard.louisbourgault.com)
Web editor for the micropython sdk: [editor.louisbourgault.com](https://editor.louisbourgault.com)

## Project Structure
- /software - all software to run both on the machine and for development
- /mini-games - The kicad project and files for the v1 hardware
- /hardware-v2 - The kicad project and files for the v2 hardware
- /case - files for the making of the case. This includes f3d and stl files.
- /img - the images used in this readme
- /mp-sdk - the legacy micropython sdk, and setup to build to a custom uf2. More information on this system can be found in the readme in this folder.
- /web-editor - a fully featured web based editor for programs written in the micropython sdk, communicating over webserial.

Production files: production files can be found at /jlcpcb/production-files in both the mini-games and hardware-v2 directories, for the v1 and v2 design respectively. These are automatically generated with the jlcpcb tools plugin in kicad, and use LCSC numbers for BOM and CPL. 

# Assembly instructions
Assembly instructions can be found at [ferretboard.louisbourgault.com/docs/assembly](https://ferretboard.louisbourgault.com/docs/assembly).

# Gallery

The full gallery can be found at [ferretboard.louisbourgault.com/gallery](https://ferretboard.louisbourgault.com/gallery).
## The micropython online editor working!
![micropython image system](/img/micropython-images.jpg)

## Finished Device (v1)
![Finished device](/img/finished-device.jpg)

## Fusion design for new version (not produced yet)
![New Version](/img/newcase.png)

# Acknowledgements

Find all acknowledgements at [ferretboard.louisbourgault.com/thanks](https://ferretboard.louisbourgault.com/thanks)

Font file used in /software/helpers/text.go - Public Domain, [github.com/dhepper/font8x8](https://github.com/dhepper/font8x8/blob/master/font8x8_basic.h)

framebuf2 - MIT Licenced, https://github.com/peter-l5/framebuf2

codemirror v6 - MIT Licenced, https://code.haverbeke.berlin/codemirror/dev/ 

# Licensing

This project is dual-licensed:
* **Hardware:** The schematic, PCB layout, and documentation files are licensed under the [CERN Open Hardware Licence v2 - Strongly Reciprocal (CERN-OHL-S-2.0)](./mini-games/LICENSE.txt). If you modify and distribute this hardware, you must share your changes to it. 
* **Firmware/Software:** All source code is licensed under the [GNU AGPLv3](./LICENSE.txt). You can't copy and distribute a changed version of this without making your modifications public, including distribution by serving a part of this code over a network. [Even if you're Bambu Labs.](https://consumerrights.wiki/w/Bambu_Lab_cease_and_desist_against_OrcaSlicer_fork_developer)

# Looking for something?
Looking for something, like assembly instructions or something else? Please check the link up the top of this readme. 