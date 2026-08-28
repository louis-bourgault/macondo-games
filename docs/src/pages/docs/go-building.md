---
layout: '../../layouts/DocsLayout.astro'
title: 'Building the Go system'
draft: false
category: "go"
---

This page is for the standalone Go games system. If you want to use the current MicroPython SDK and web editor, follow the [MicroPython building guide](/docs/micropython-building) instead.

### Wasm (golang system on web)
You'll need to compile the binary for wasm using tinygo. I'm on a mac, so keep in mind that all these commands are macos specific and may be different for you. For me, tinygo is installed through Homebrew.

Run these commands from the ```/software``` directory:

```sh
GOOS=js GOARCH=wasm tinygo build -o ./web/main.wasm -target wasm -tags=wasm ./cmd/wasm
cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" ./web/
```

It can be quicker to use the typical Go WASM handler for builds, since the tinygo toolchain can take quite a while, even on a decent computer. In this case, the js file is different, so the commands are:

```sh
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./web/
GOOS=js GOARCH=wasm go build -o ./web/main.wasm ./cmd/wasm
```

Then, regardless of which you use, you need to actually serve this directory. For debugging, I do this through cd'ing into it and then running ```python3 -m http.server```, which works well enough for me.

If you're building the version used by the Astro frontend, run these commands from ```/software```:

```sh
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../docs/public/
GOOS=js GOARCH=wasm go build -o ../docs/public/main.wasm ./cmd/wasm
```

### Building (golang system on device)
To compile the standalone Go games firmware for the version one RP2040 device, you'll need tinygo installed on your computer. Then, cd into the ```/software``` folder and run this command:

```sh
tinygo build -target=pico -o ./out/firmware.uf2 ./cmd/device
```

This will create a .uf2 build artifact in the ```/software/out``` directory. Alternatively, I have chosen not to include this in my .gitignore, so you can find a prebuilt file in this repository at that location, although i make no guarantees about it being up to date.
