# Web Editor
This repository is a feature complete web editor for the python sdk of the board, which includes a multifile code editor (based on the excellent codemirror library), image sprite editor with transparency, and terminal. The system connects to the device over USB using the browser WebSerial API.

It also includes an RP2350 browser emulator which runs MicroPython in a Web
Worker and implements the same `ferret` API against a canvas and ABXY controls.
See [EMULATOR.md](./EMULATOR.md) for its architecture, parity rules and hosting
headers.

## Stack

This web app uses a simple typescript setup comprising of:
- Svelte
- Sveltekit
- UI library: shadcn-svelte
- codemirror

There is no database or backend component, and this can be entirely run statically by building sveltekit to the static adapter.

To program a device using this sdk, you must have a device flashed with the micropython sdk, which can be found at /mp-sdk at the top level of this repository. Follow the instructions in that README for instructions on building and flashing this firmware.

## Running locally

Install modules and run:

```bash
pnpm install

pnpm run dev
```

Then open your browser to localhost:5173 to test.

## Requirements
The web serial api is not supported by Safari or Firefox. Use a recent version of a chromium based browser.

## Licence
As with all other code components in this repository, this module is licenced under the [GNU AGPL](../LICENSE.txt)


