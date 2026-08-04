![banner](/img/banner.png)
# ferretboard!!!

> [!IMPORTANT]
> # PLEASE PLEASE READ IM ACTUALLY BEGGING YOU
> The primary documentation for this project can be found at [ferretboard.louisbourgault.com/docs](https://ferretboard.louisbourgault.com/docs). Please read from that source for organised information on software interfaces, assembly, and a more complete project overview.
> 
> This README remains the authoritative source on commands to build the firmware for the device, and also has a nice image gallery at the end.


### a small games console with rp2040
This is a simple games console, based around a custom PCB with the RP2040 chip and a 240x240 LCD screen. It's small, of dimensions about 50x100x15mm

This is made as a part of Hack Club Macondo. You can find my project at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137)

## Online stuff!
Docs site and main site: [ferretboard.louisbourgault.com](https://ferretboard.louisbourgault.com)
Web editor for the micropython sdk: [editor.louisbourgault.com](https://editor.louisbourgault.com)

## Project Structure
- /software - all software to run both on the machine and for development
- /mini-games - The kicad project and files for the hardware
- /mini-games/jlcpcb/production_files - The files for sending to JLC, exported by the jlcpcb tools plugin in kicad. Big thanks to this plugin: [Bouni/kicad-jlcpcb-tools](https://github.com/bouni/kicad-jlcpcb-tools)
  - These are also duplicated in /production, for ease of review.
- /case - files for the making of the case. This includes f3d and stl files.
- /img - the images used in this readme
- /mp-sdk - the micropython sdk, and setup to build to a custom uf2. More information on this system can be found in the readme in this folder.
- /web-editor - a fully featured web based editor for programs written in the micropython sdk, communicating over webserial.

## Building (Go system)
### Wasm (golang system on web)
You'll need to compile the binary for wasm using tinygo. I'm on a mac, so keep in mind that all these commands are macos specific and you may be different. For me, tinygo is installed through Homebrew.

```GOOS=js GOARCH=wasm tinygo build -o ./software/web/main.wasm -target wasm -tags=wasm ./software/cmd/wasm```

```cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js /software/web/```


It can be quicker to use the typical go WASM handler for builds, since the tinygo toolchain can take quite a while, even on a decent computer. In this case, the js file is different, so the command is:


Full command: ```cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./web/ && GOOS=js GOARCH=wasm go build -o ./web/main.wasm ./cmd/wasm``` (you must be cded into the software directory for this to work)

Then, regardless of which you use, you need to actually serve this directory. For debugging, I do this through cd'ing into it and then running ```python3 -m http.server```, which works well enough for me. This is also the directory that I serve on Vercel.

or if you're building for the new astro frontend, you can run ```cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../docs/public/ && GOOS=js GOARCH=wasm go build -o ../docs/public/main.wasm ./cmd/wasm```

### Building (golang system on device)
To compile for the actual device, you'll need tinygo installed on your computer. Then, cd into the ```/software``` folder and run this command:

```tinygo build -target=pico -o ./out/firmware.uf2 ./cmd/device```

This will create a .uf2 build artifact in the ```/software/out``` directory. Alternatively, I have chosen not to include this in my .gitignore, so you can find a prebuilt file in this repository at that location, although i make no guarantees about it being up to date.

### SDK information
You can find sdk information and function descriptions at the docs site linked at the top of this README.

# Micropython SDK
In addition to the Go programming system for this device, there is a micropython sdk for this device, intended for ease of development and simplicity.
The SDK for this can be found in the ```/mp-sdk``` directory in this repository. In addition, there is a web editor for this micropython sdk, which can be found in ```/web-editor```.

This micropython sdk was created for ease of learning to code this system. I was talking to a digital technologies teacher at my school, showing him this board, and he expressed interest in how this kind of system might work for a game development class, because its a way to teach people to program with immediate feedback, and less annoying boilerplate than learning game dev through Unity, Godot et cetera.

## Building the micropython sdk.
To build the micropython sdk, follow instructions in [its readme](/mp-sdk/README.md)

# Assembly instructions
Assembly instructions can be found at [ferretboard.louisbourgault.com/docs/assembly](https://ferretboard.louisbourgault.com/docs/assembly).

# Image Gallery
![Fusion Case](/img/fusion-case.png)
![With Parts](/img/fusion-with-parts.png)
![Schematic](/img/Schematic.png)

## JLC pcb render and cart
![JLC Cart](/img/JLC-Cart.png)
![top of pcb](/img/JLC-Topside.png)

## The website that hosts the WASM version and documentation.
[ferretboard.louisbourgault.com](https://ferretboard.louisbourgault.com)
![Web Screenshot](/img/web-version.png)

## The micropython editor working!
![micropython image system](/img/micropython-images.jpg)

## Finished Device
![Finished device](/img/finished-device.jpg)

# Acknowledgements
Font file used in /software/helpers/text.go - Public Domain, [github.com/dhepper/font8x8](https://github.com/dhepper/font8x8/blob/master/font8x8_basic.h)

framebuf2 - MIT Licenced, https://github.com/peter-l5/framebuf2

codemirror v6 - MIT Licenced, https://code.haverbeke.berlin/codemirror/dev/ 


# Things that I'd change if i did it again
- perhaps something like a ground plane on the pcb
- remove resistor that stops the reset button from working
- make it possible to charge the device while it is off
- make the on/off switch in an easier to reach location


# Licensing

This project is dual-licensed:
* **Hardware:** The schematic, PCB layout, and documentation files are licensed under the [CERN Open Hardware Licence v2 - Strongly Reciprocal (CERN-OHL-S-2.0)](./mini-games/LICENSE.txt). If you modify and distribute this hardware, you must share your changes to it. 
* **Firmware/Software:** All source code is licensed under the [GNU AGPLv3](./LICENSE.txt). You can't copy and distribute a changed version of this without making your modifications public, including distribution by serving a part of this code over a network. [Even if you're Bambu Labs.](https://consumerrights.wiki/w/Bambu_Lab_cease_and_desist_against_OrcaSlicer_fork_developer)

# Looking for something
Looking for something, like assembly instructions or something else? Please check the link up the top of the 