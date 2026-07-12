---
layout: '../../layouts/DocsLayout.astro'
title: 'Project Structure'
draft: false
category: "general"
---

## File structure

## Project Structure
- /software - all software to do with the golang system
- /mini-games - The kicad project and files for the hardware
- /mini-games/jlcpcb/production_files - The files for sending to JLC, exported by the jlcpcb tools plugin in kicad. Big thanks to this plugin: [Bouni/kicad-jlcpcb-tools](https://github.com/bouni/kicad-jlcpcb-tools)
  - These are also duplicated in /production, for ease of review.
- /case - files for the making of the case. This includes f3d and stl files.
- /img - the images used in this readme
- /docs - this website
- /mp-sdk - the micropython sdk
- /web-editor - the web editor for the micropython sdk.

