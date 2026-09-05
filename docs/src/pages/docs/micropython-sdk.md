---
layout: '../../layouts/DocsLayout.astro'
title: 'MicroPython SDK (new) - About'
draft: false
category: "mp"
---

The current MicroPython SDK is the main scripting system for the console. You write your code in the web editor at ```/web-editor``` in the repository, which is deployed at [editor.louisbourgault.com](https://editor.louisbourgault.com).

The intention of this system is to provide an easy entry to the console. I am first writing this with my idea being to give these devices to a school as a way to teach students to code, or at least to pitch the idea to a teacher. I think that having an easy to code system that runs fully in a browser without having to install anything is pretty cool, personally.

This is different to the [legacy MicroPython SDK](/docs/micropython-sdk-legacy). The current firmware has a Go backend which controls the display, input and filesystem, with MicroPython embedded into it and a thin C glue layer in the middle. Student programs use the ```ferret``` module to call into that backend.

## Using the editor

The editor uses Web Serial, so you need a recent Chromium-based browser such as Chrome or Edge, and the page must be loaded over HTTPS or localhost. Safari and Firefox do not currently support Web Serial.

Every project has a ```main.py``` file, which is run automatically when the device starts or soft resets. Other ```.py``` files can be imported from ```main.py```. Their names must be valid Python identifiers followed by ```.py``` (for example, ```player.py```). Start programs by importing the SDK:

```python
import ferret
```

The editor can create multiple code files, draw and import images, show the serial terminal, save a project in browser storage, and download or load the complete project as a JSON file. Connect the device, then use **Run Script** to synchronize the project's code files to flash (including removing files deleted or renamed in the editor) and soft reset into the new ```main.py```. Images are sent separately with **Sync Images**.

See the [embedded reference](/docs/embed-reference) for the functions available to games, and the [building guide](/docs/micropython-building) for making the firmware itself.

## Included Python modules

The firmware intentionally includes a fairly small set of Python modules. The most useful ones for games are ```time```, ```math```, ```struct```, ```array```, ```collections```, ```gc``` and ```sys```. Normal Python filesystem functions such as ```open()``` and ```os.listdir()``` are not included in the current system; code and images are managed by the Go backend and the web editor.

![images in the sdk, working on the device](https://cdn.hackclub.com/01a047fd-d714-773d-8a16-4d080555ccab/micropython-images.jpg.webp)
