# Macondo Games
I would like to make a simple games console, based around an RP2040 chip on a custom PCB, and including a 240x240 LCD screen. I am designing this roughly to fit in the footprint of an iPod nano shape and size - hopefully acheiving final dimensions of about 550x1000x100mm.

## Components
The project will be made up of a custom pcb, including the chip, 16MB flash, battery management circuits etc. I will also use:
- 3.7v 1000mAh LiPo battery
- 1.54" LCD Screen (Already purchased) of pixel dimensions 240x240 and through an SPI interface

The exterior will be 3d printed

## Controls
D-pad controls (Up, down, left, right), as well as A and B buttons and 2 buttons for other functions, such as going to menus, pausing games, etc

## Games
I initially considered making this a gameboy emulator, but decided that that was against the spirit of the project, since I would like to code most of the things on there and don't particularly feel like playing gameboy games without sound. Initially, I will code games like Pong and Snake in Go, and then who knows what I will go to past then.
The one exception to trying to code all my own games is that, as a rite of passage into hardware engineering, I want to run Doom. 

For development, another thing that i want to do is to have all the interfaces for inputs, drawing to the screen, etc, abstracted. Then I can use build tags to either build the software using TinyGo for loading onto the device or using normal Go for debugging on my computer, using a window on my screen and keyboard controls. It would also be cool if I could build it into WASM, so that there can be an online demo for the purposes of Hack Club - this could even be in replacement of the local native running on my device for debugging.

## Project Structure
- /software - all software to run both on the machine and for development
- /mini-games - The kicad project and files for the hardware
- /mini-games/jlcpcb/production_files - The files for sending to JLC, exported by the jlcpcb tools plugin in kicad. Big thanks to this plugin: [Bouni/kicad-jlcpcb-tools](https://github.com/bouni/kicad-jlcpcb-tools)


## Building
### Wasm (web)
You'll need to compile the binary for wasm using tinygo. I'm on a mac, so keep in mind that all these commands are macos specific and you may be different. For me, tinygo is installed through Homebrew.

```GOOS=js GOARCH=wasm tinygo build -o ./software/web/main.wasm -target wasm -tags=wasm ./software/cmd/wasm```

```cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js /software/web/```


It can be quicker to use the typical go WASM handler for builds, since the tinygo toolchain can take quite a while, even on a decent computer. In this case, the js file is different, so the commands are:
```cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./software/web/```
```GOOS=js GOARCH=wasm go build -o software/web/main.wasm ./software/cmd/wasm```

Then, regardless of which you use, you need to actually serve this directory. For debugging, I do this through cd'ing into it and then running ```python3 -m http.server```, which works well enough for me


# TODOS
- implement device display logic
- implement clock with framerate 
- implement input system for both device and wasm
- make some actually cooler games