#!/bin/zsh

#Note: this is for my personal use, and has hardcoded paths to my local repositories. 
#If you want to use it, change this to point to where everything is on your machine
#If you're on linux, use bash instead of zsh (should work the same, i just use zsh because im on mac)
#also remember to chmod +x build.sh so that this script actually runs

cd "/Users/louisb/Documents/hardware/micropython/ports/rp2"; rm -rf build-ferretboard; make BOARD_DIR="/Users/louisb/Documents/hardware/macondo-games/mp-sdk/boards/ferretboard" FROZEN_MANIFEST="/Users/louisb/Documents/hardware/macondo-games/mp-sdk/boards/ferretboard/manifest.py"
cp "/Users/louisb/Documents/hardware/micropython/ports/rp2/build-ferretboard/firmware.uf2" "/Users/louisb/Documents/hardware/macondo-games/mp-sdk/firmware.uf2"