# uploading to the device
it doesn't work just hitting run in vs code: hit all commands and then do micropico: upload project to pico from command/control shift p


# Building instructions:
## Setup - what you need installed on your computer (macOS, likely very similar on linux, if you're on windows use WSL)
If you're on linux, all you'll need to change is to install the packages below with your distro's package manager

```zsh
brew install --cask gcc-arm-embedded
brew install cmake ninja
```

next, clone the micropython git repository and compile some things

```zsh
git clone https://github.com/micropython/micropython.git
cd micropython
make -C mpy-cross

cd ports/rp2
make submodules
```

## Building
Hop into the ```ports/rp2``` directory inside your micropython git clone.

```zsh
# delete old build cache
rm -rf build-ferretboard

#run the build command with paths to the board directory and the manifest.py. For me this is
make BOARD_DIR=/Users/louisb/Documents/hardware/macondo-games/mp-sdk/boards/ferretboard FROZEN_MANIFEST=/Users/louisb/Documents/hardware/macondo-games/mp-sdk/modules/manifest.py

```

This will output a uf2 file in the ```ports/rp2/build-ferretboard/``` directory inside the micropython repo, which you can then copy onto the device