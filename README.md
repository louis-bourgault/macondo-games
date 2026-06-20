![banner](/img/banner.png)
# ferretboard!!!

> [!IMPORTANT]
> The primary documentation for this project can be found at [ferretboard.louisbourgault.com/docs](https://ferretboard.louisbourgault.com/docs). Please read from that source for project information

### a small games console with rp2040
I would like to make a simple games console, based around an RP2040 chip on a custom PCB, and including a 240x240 LCD screen. I am designing this roughly to fit in the footprint of an iPod nano shape and size. The overall dimensions are around 50x100x10mm
This is part of Hack Club Macondo. You can find my project at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137)

## Components
The project will be made up of a custom pcb, including the chip, 16MB flash, battery management circuits etc. I will also use:
- 3.7v 1000mAh LiPo battery
- 1.54" LCD Screen (Already purchased) of pixel dimensions 240x240 and through an SPI interface
- 6 basic 6x6x5mm push to make switches (i had some lying around so decided not to get jlc to assemble them for me.)

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

## Hardware Constraints
This project is shaped by its hardware constraints of the device I designed. The RP2040 runs on a 135MHz dual core CPU, and has 264KB of ram (115 of which are taken up by the display buffer). Therefore, we have to be efficient.
For flash, the hardware has a 16mb flash chip, which is massive. Realistically, there's no way we're filling that up unless i include a ton of full res background images.

## Software Interfaces
The system is designed to be as interchangeable as possible. There are interfaces defined in ```/software/internal/platform/platform.go```, which describe the shape of the input and display handlers for both the wasm system and the embedded system. Games are also fully abstracted: there's an interface at ```/software/internal/game/game.go``` which describes the shape of the games. 

Basically, every game defines three functions: a New() function, which is referenced in ```/software/internal/game/menu/menu.go``` which takes no arguments and returns an initialised game, an update function that takes a reference to the input system and delta time and returns a game (if you want to quit the game, you return nil and the manager system will automatically send the user back to the menu, otherwise just return the current game), and a draw function which takes a reference to the screen object being used, which you can draw to.

Each of the screen interfaces provides basic, optimised functions for filling the entire screen, drawing a rectangle, and setting an individual pixel. As well as that, there are a bunch of functions in ```/software/internal/helpers/``` to do with rendering text (i will add more later) that in turn act by just calling the Pixel() function a ton of times.

If you really want to contribute, i wouldn't mind if you made a game using the interfaces, and just chucked it in its own folder in the /internal/game/ directory. As long as you don't change anything other than that directory and the menu entry to instantiate it, i'll probably merge it. Only if you want to, though. And make sure you use the hardware efficiently.

### Images
to include an image in the program, you'll need to process it before you can use it with DrawSprite. First, draw your image in whatever program you use, like krita or ms paint. Then, put your image into /helpers, rename it to image.png, and run the python script. 
The python script will give you a .bin file that you can put into your code with //go:embed. For example, in the flappy bird game, there's a system.
I also made a website at /software/web/convert that does the same thing, with a nice ui. This is fully vibe coded in like 30 seconds, but works ok for what it needs to. It's just a simple tool, i didn't want to devote any actual coding time to.

```go
//go:embed bird.bin
var rawBirdData string

var BirdSprite = helpers.Image{
	Data: rawBirdData,
	W:    12,
	H:    12,
}
```
This means that the data of the sprite never clogs up our ram, and is streamed straight from the flash ROM to the screen buffer!

## Building
### Wasm (web)
You'll need to compile the binary for wasm using tinygo. I'm on a mac, so keep in mind that all these commands are macos specific and you may be different. For me, tinygo is installed through Homebrew.

```GOOS=js GOARCH=wasm tinygo build -o ./software/web/main.wasm -target wasm -tags=wasm ./software/cmd/wasm```

```cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js /software/web/```


It can be quicker to use the typical go WASM handler for builds, since the tinygo toolchain can take quite a while, even on a decent computer. In this case, the js file is different, so the command is:


Full command: ```cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./web/ && GOOS=js GOARCH=wasm go build -o ./web/main.wasm ./cmd/wasm``` (you must be cded into the software directory for this to work)

Then, regardless of which you use, you need to actually serve this directory. For debugging, I do this through cd'ing into it and then running ```python3 -m http.server```, which works well enough for me. This is also the directory that I serve on Vercel.

or if you're building for the new astro frontend, you can run ```cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../docs/public/ && GOOS=js GOARCH=wasm go build -o ../docs/public/main.wasm ./cmd/wasm```

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

## Finished Device
![Finished device](/img/finished-device.jpg)

# Acknowledgements
Font file used in /software/helpers/text.go: [github.com/dhepper/font8x8](https://github.com/dhepper/font8x8/blob/master/font8x8_basic.h)

# Things that I'd change if i did it again
- perhaps something like a ground plane on the pcb
- remove resistor that stops the reset button from working