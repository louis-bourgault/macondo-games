---
layout: '../../layouts/DocsLayout.astro'
title: 'Changes for V2'
draft: false
category: "general"
---

The v2 of this hardware is significantly different to version one. Here is a description of what these differences are, why they are, and what this changes for the device

## Change from rp2040 to rp2350
This is the main change that motivated the redesign in the first place. The rp2040 has only 264KB of SRAM, which is fine, but we are starting to outgrow it. This limits us in quite a few ways, including limiting us to a single buffer rather than a double buffer. It also runs the risk of OOM errors, especially when we're dealing with the embedded micropython system - this introduces the overhead of both micropython and go. So far, this is ok, but struggling, on the rp2040 - there have been some OOM problems with uploading images to the device and displaying them from the editor. The rp2350 upgrades to 520KB of internal ram, which gives us room for

- a double buffer (230kb)
- a cache for recently used sprites, which includes a full screen background and some smaller sprites. the full screen background here will be another 115kb of image, and then a budget of perhaps 64kb or so for sprites. That sums up to 179KB there.

After this image caching, we get 110KB of headroom for the Go process, micropython, and everything else. I think that should be enough - it is still cutting it a bit fine, but should be good.

Apart from the RAM, the RP2350 also gives us more processing power. The RP2040 uses a dual core ARM CORTEX M0+ at 133MHz - the 2350 has dual core CORTEX M33 at 150MHz. This improves speeds by at least double based on a lot of stuff I could find online. Also, the 2350 has a designated FPU, so the maths simulations that struggled so much on the 2040 might actually run.

## New Form Factor
I still really like the form factor of the original. That being said, its not very ergonomic, and is kind of cramped to have both thumbs in the tiny space that it has. V2 changes this to a more traditional setup with d-pad on the left, the screen in the middle, and ABXY on the right. This should feel better to hold.

The basic buttons on v1 are also replaced with softer touch buttons that feel more like a games controller on v2 (they feel VERY similar to the pushing force required on my Dualsense controller). Will be nice.

## Minor things and general polish
A few things here - the case is better designed to fit around it, and i think will look really nice. There's also a few things fixed - the reset button on v1 didn't work, meaning that to reset the thing at the moment i have to manually short QSPI_SS to ground with tweezers while rebooting it. This is obviously not ideal.

Rotation was annoying on v1 - the st7789 chip in the display is meant for a 320x240 display, and try as i might i couldnt get the rotation parameters to work on it; thus, i had to do software rotation for EVERYTHING. I haven't benchmarked it, but i imagine this isnt great for performance. The display is mounted the right way up, not upside down, in v2, so the native coordinates work without rotation.