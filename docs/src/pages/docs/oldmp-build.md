---
layout: '../../layouts/DocsLayout.astro'
title: 'Building the old micropython system'
draft: false
category: "ogmp"
---

These are instructions for building the legacy micropython games system. There's really no reason for you to use it - its kind of buggy, and just bad compared to the new one. 

The current deployed editor is no longer compatible with this old system. You can find the old editor deployment here: [https://editor-puhp0agjy-louisbs-projects.vercel.app?_vercel_share=TvlfneAJnYLfwqdn5NOvuiplwbXqOOF9](https://editor-puhp0agjy-louisbs-projects.vercel.app?_vercel_share=TvlfneAJnYLfwqdn5NOvuiplwbXqOOF9). This deployment is pinned to a commit before the editor changes happened. Alternatively, you can roll back to git commit **5ab1194** on the main branch and serve the Sveltekit server locally.

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

Note: once you're at this stage, you can use the script ```build.sh``` in this repository to build it. Note, you will have to adapt this script to where everything is on your machine, so its probably easier to just do the commands manually.

## Building
Hop into the ```ports/rp2``` directory inside your micropython git clone.

```zsh
# delete old build cache
rm -rf build-ferretboard

#run the build command with paths to the board directory and the manifest.py. For me this is
make BOARD_DIR=/Users/louisb/Documents/hardware/macondo-games/mp-sdk/boards/ferretboard FROZEN_MANIFEST=/Users/louisb/Documents/hardware/macondo-games/mp-sdk/boards/ferretboard/manifest.py

```

This will output a uf2 file in the ```ports/rp2/build-ferretboard/``` directory inside the micropython repo, which you can then copy onto the device. If you're using the script as i am, it will instead copy it directly into this repository.

# Licences
This software itself is AGPL licenced. 

Acknowledgement is also given to the framebuf2 library, which makes up an integral part of this repository. This component is licenced under a MIT licence. 