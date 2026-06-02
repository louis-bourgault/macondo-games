![banner](/img/banner.png)
# ferretboard!!!
### a small games console with rp2040
I would like to make a simple games console, based around an RP2040 chip on a custom PCB, and including a 240x240 LCD screen. I am designing this roughly to fit in the footprint of an iPod nano shape and size. The overall dimensions are around 50x100x10mm
This is part of Hack Club Macondo. You can find my project at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137)

## Components
The project will be made up of a custom pcb, including the chip, 16MB flash, battery management circuits etc. I will also use:
- 3.7v 1000mAh LiPo battery
- 1.54" LCD Screen (Already purchased) of pixel dimensions 240x240 and through an SPI interface

The exterior will be 3d printed. I can print these things myself at school.

## Controls
D-pad controls (Up, down, left, right), as well as A and B buttons and 2 buttons for other functions, such as going to menus, pausing games, etc

## Games
I initially considered making this a gameboy emulator, but decided that that was against the spirit of the project, since I would like to code most of the things on there and don't particularly feel like playing gameboy games without sound. Initially, I will code games like Pong and Snake in Go, and then who knows what I will go to past then.
The one exception to trying to code all my own games is that, as a rite of passage into hardware engineering, I want to run Doom.

The development is well abstracted, so that the game logic isn't actually connected to any of the code for the SPI interface, for example. There is an interface for the device inputs and screen outputs, and a seperate interface for a WASM build that uses the keyboard to control it and renders to a html canvas. You can find this online:
[ferretboard.louisbourgault.com](https://ferretboard.louisbourgault.com)

## Project Structure
- /software - all software to run both on the machine and for development
- /mini-games - The kicad project and files for the hardware
- /mini-games/jlcpcb/production_files - The files for sending to JLC, exported by the jlcpcb tools plugin in kicad. Big thanks to this plugin: [Bouni/kicad-jlcpcb-tools](https://github.com/bouni/kicad-jlcpcb-tools)
  - These are also duplicated in /production, for ease of review.
- /case - files for the making of the case. This includes f3d and stl files.
- /img - the images used in this readme


## Building
### Wasm (web)
You'll need to compile the binary for wasm using tinygo. I'm on a mac, so keep in mind that all these commands are macos specific and you may be different. For me, tinygo is installed through Homebrew.

```GOOS=js GOARCH=wasm tinygo build -o ./software/web/main.wasm -target wasm -tags=wasm ./software/cmd/wasm```

```cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js /software/web/```


It can be quicker to use the typical go WASM handler for builds, since the tinygo toolchain can take quite a while, even on a decent computer. In this case, the js file is different, so the command is:


Full command: ```cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./web/ && GOOS=js GOARCH=wasm go build -o ./web/main.wasm ./cmd/wasm && cd ..``` (you must be cded into the software directory for this to work)

Then, regardless of which you use, you need to actually serve this directory. For debugging, I do this through cd'ing into it and then running ```python3 -m http.server```, which works well enough for me.

### Building (device)
You'll need to compile it with TinyGo, as well. I'll add instructions for this once I have the device built.
This is something along the lines of ```tinygo build -target=pico -o ./out/firmware.uf2 ./cmd/device```

# Image Gallery
![Fusion Case](/img/fusion-case.png)
![With Parts](/img/fusion-with-parts.png)
![Schematic](/img/Schematic.png)

## JLC pcb render and cart
![JLC Cart](/img/JLC-Cart.png)
![top of pcb](/img/JLC-Topside.png)

## The website that hosts the WASM version
[ferretboard.louisbourgault.com](https://ferretboard.louisbourgault.com)
![Web Screenshot](/img/web-version.png)