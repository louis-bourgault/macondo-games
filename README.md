![banner](/img/banner.png)
# ferretboard!!!

> [!IMPORTANT]
> The primary documentation for the software system of this project can be found at [ferretboard.louisbourgault.com/docs](https://ferretboard.louisbourgault.com/docs). Please read from that source for better organised information on software interfaces, etc.
> 
> This README remains the authoritative source on commands to build the firmware for the device.

### a small games console with rp2040
I would like to make a simple games console, based around an RP2040 chip on a custom PCB, and including a 240x240 LCD screen. I am designing this roughly to fit in the footprint of an iPod nano shape and size. The overall dimensions are around 50x100x10mm
This is part of Hack Club Macondo. You can find my project at [macondo.hackclub.com/projects/1137](https://macondo.hackclub.com/projects/1137)

## Components
The project will be made up of a custom pcb, including the chip, 16MB flash, battery management circuits etc. I will also use:
- 3.7v 1000mAh LiPo battery
- 1.54" LCD Screen (Already purchased) of pixel dimensions 240x240 and through an SPI interface. Must use the ST7789 driver chip to work without code changes.
- 6 basic 6x6x5mm push to make switches (i had some lying around so decided not to get jlc to assemble them for me.)
- 3d printed case, the stl file for which is in ```/case/case.stl``` in this repository


## Controls
D-pad controls (Up, down, left, right), as well as A and B buttons and 2 buttons for other functions, such as going to menus, pausing games, etc

## Assembly instructions
To assemble the device, start by 3d printing the shell, which can be found at ```/case/case.stl``` in this repository. I would reccomend doing this with some kind of high precision setting on your printer - my first print used the Draft quality with layer height of 0.2mm, but my second used 0.08mm and it looked a lot better. I just printed this on my school's Bambu Lab P1S printer. Here's how it looks, printed in white on this 0.08mm setting.
![The case, printed in white](/img/whitecase.jpg)
Next, you'll need to solder the buttons to the PCB that you've got printed and assembled. Get 6 generic push button push to make switches, and solder them into the holes on the pcb. Get the battery (no smaller than 500mAh), strip the JST connector off it if it has one, and solder the positive and negative lines to the copper pads of the board as indicated below:
![Where to solder things](/img/instructions.jpg)
Key: Battery as indicated, positive to J4, negative to J3. Make sure these lines do not touch.
Switches: solder legs to tht holes for SW2-SW7, as highlighted in yellow.
Next, get your screen and solder it to the screen connector at the top of the PCB. The leftmost pin, from the front of the board, is the BL pin, with GND on the far right. This means that I had to mount the specific display I bought from aliexpress upside down, and flip all the coordinates in software.
![Connecting the screen](/img/screensolder.jpg)
Once you've done that, the device is pretty much ready to go. Slot the whole contraption into the 3d printed case - the case is designed so that the battery goes behind the screen and then the wires for the battery come down and snake through the little notch to the left of the battery compartment. There are four posts on the lower section of the case, which you can use to hold the PCB in place through the circular holes on each corner of the PCB. 
If you leave it like that, it will work, but depending on the sizes of all your components, everything will probably rattle and tilt and slide around a ton. I solved this problem by liberally applying small bits of furniture skates, double sided tape, and foam to hold everything in place. This is something that I can probably improve on in version two of the system.
From there, all you'll need to do is to load the code onto the thing. Plug it in via USB C to your computer, and drag the compiled UF2 file from the compilation instructions below onto the drive that shows up on your file explorer. It'll reboot into the menu, and you're ready to go.

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
To compile for the actual device, you'll need tinygo installed on your computer. Then, cd into the ```/software``` folder and run this command:

```tinygo build -target=pico -o ./out/firmware.uf2 ./cmd/device```

This will create a .uf2 build artifact in the ```/software/out``` directory. Alternatively, I have chosen not to include this in my .gitignore, so you can find a prebuilt file in this repository at that location, although i make no guarantees about it being up to date.

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
- make it possible to charge the device while it is off
- make the on/off switch in an easier to reach location


## Licensing

This project is dual-licensed:
* **Hardware:** The schematic, PCB layout, and documentation files are licensed under the [CERN Open Hardware Licence v2 - Permissive (CERN-OHL-P)](./mini-games/LICENSE.txt).
* **Firmware/Software:** All source code is licensed under the [MIT License](./LICENSE).
