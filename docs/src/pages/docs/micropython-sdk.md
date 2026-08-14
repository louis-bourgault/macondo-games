---
layout: '../../layouts/DocsLayout.astro'
title: 'Micropython SDK (new) - About'
draft: false
category: "ogmp"
---

note: this docs page is a wip


The micropython sdk is the new scripting language for the system - you write your code in the web editor at ```/web-editor``` in the repository (deployed at [https://editor.louisbourgault.com](editor.louisbourgault.com).)

The intention of this system is to provide an easy entry to the system. I am first writing this with my idea being to give these devices to school as a way to teach students to code, or at least to pitch the idea to a teacher. I think that having an easy to code system that runs fully in a browser without having to install anything is pretty cool, personally.

This is better than the other micropython sdk because its backend is in go, and the micropython is just a thin wrapper on top, with a thin C glue layer in the middle. This is because I like coding in go more, and i can get more performance from go theoretically (untested, but i dont see why not).
