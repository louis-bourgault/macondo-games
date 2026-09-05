/// <reference lib="webworker" />

import { loadMicroPython } from '@micropython/micropython-webassembly-pyscript/micropython.mjs';
import micropythonWasmUrl from '@micropython/micropython-webassembly-pyscript/micropython.wasm?url';
import { FONT_8X8 } from './font.generated';
import type { FerretApiName } from './api.generated';
import { BUTTONS, SCREEN_SIZE, buttonIndex, type EmulatorButton, type EmulatorCommand, type EmulatorEvent } from './protocol';
import type { ImageFile, ProjectData } from '$lib/types';

const ctx: DedicatedWorkerGlobalScope = self as unknown as DedicatedWorkerGlobalScope;
const CHROMA_KEY = 0xf81f;
const framebuffer = new Uint16Array(SCREEN_SIZE * SCREEN_SIZE);
type StoredImage = { width: number; height: number; pixels: Uint16Array; base64: string };
let images = new Map<string, StoredImage>();
let imageUpload: { name: string; width: number; height: number; base64: string } | null = null;
let physicalButtons: Int32Array | null = null;
let sampledButtons = 0;
let previousButtons = 0;
let micropython: any;

function emit(event: EmulatorEvent, transfer: Transferable[] = []) {
	ctx.postMessage(event, transfer);
}

function pixel(x: number, y: number, color: number) {
	if (color === CHROMA_KEY || x < 0 || y < 0 || x >= SCREEN_SIZE || y >= SCREEN_SIZE) return;
	framebuffer[y * SCREEN_SIZE + x] = color & 0xffff;
}

function loadImages(files: ImageFile[]) {
	images = new Map();
	for (const image of files) {
		const binary = atob(image.content);
		const pixels = new Uint16Array(image.width * image.height);
		for (let i = 0; i < pixels.length; i++) {
			pixels[i] = binary.charCodeAt(i * 2) | (binary.charCodeAt(i * 2 + 1) << 8);
		}
		images.set(image.name, { width: image.width, height: image.height, pixels, base64: image.content });
	}
}

function decodeBase64Image(name: string, width: number, height: number, base64: string): StoredImage {
	const binary = atob(base64);
	if (binary.length !== width * height * 2) throw new Error(`image data size mismatch: ${name}`);
	const pixels = new Uint16Array(width * height);
	for (let i = 0; i < pixels.length; i++) pixels[i] = binary.charCodeAt(i * 2) | (binary.charCodeAt(i * 2 + 1) << 8);
	return { width, height, pixels, base64 };
}

function checksum(value: string) {
	let hash = 0x811c9dc5;
	for (let i = 0; i < value.length; i++) hash = Math.imul((hash ^ value.charCodeAt(i)) >>> 0, 0x01000193);
	return (hash >>> 0).toString(16).padStart(8, '0');
}

function projectPath(path: string) {
	const name = path.replace(/^\/+/, '');
	return safePath(name);
}

function currentButtonMask() {
	let mask = 0;
	if (!physicalButtons) return mask;
	for (let i = 0; i < BUTTONS.length; i++) if (Atomics.load(physicalButtons, i)) mask |= 1 << i;
	return mask;
}

function buttonMask(button: string) {
	const index = buttonIndex(button as EmulatorButton);
	return index < 0 ? 0 : 1 << index;
}

const ferret = {
	fill(color: number) {
		framebuffer.fill(color & 0xffff);
	},
	fill_rect(x: number, y: number, width: number, height: number, color: number) {
		for (let py = y; py < y + height; py++) {
			if (py < 0 || py >= SCREEN_SIZE) continue;
			for (let px = x; px < x + width; px++) {
				if (px >= 0 && px < SCREEN_SIZE) framebuffer[py * SCREEN_SIZE + px] = color & 0xffff;
			}
		}
	},
	pixel,
	present() {
		const copy = framebuffer.slice().buffer;
		emit({ type: 'frame', pixels: copy }, [copy]);
	},
	draw_text(text: string, x: number, y: number, color: number) {
		let byteIndex = 0;
		for (const char of String(text)) {
			const code = char.codePointAt(0) ?? 0;
			if (code < 128) {
				for (let row = 0; row < 8; row++) {
					const line = FONT_8X8[code * 8 + row];
					for (let col = 0; col < 8; col++) if (line & (1 << col)) pixel(x + byteIndex * 8 + col, y + row, color);
				}
			}
			byteIndex += new TextEncoder().encode(char).length;
		}
	},
	measure_text(text: string) {
		return [new TextEncoder().encode(String(text)).length * 8, 8];
	},
	draw_image(name: string, x: number, y: number) {
		const image = images.get(name);
		if (!image) throw new Error(`image not found: ${name}`);
		for (let row = 0; row < image.height; row++) {
			for (let col = 0; col < image.width; col++) pixel(x + col, y + row, image.pixels[row * image.width + col]);
		}
	},
	rgb_to_565(r: number, g: number, b: number) {
		return (((r & 0xf8) << 8) | ((g & 0xfc) << 3) | (b >> 3)) & 0xffff;
	},
	random_int(min: number, max: number) {
		return Math.floor(Math.random() * (max - min + 1)) + min;
	},
	input_update() {
		previousButtons = sampledButtons;
		sampledButtons = currentButtonMask();
	},
	input_is_pressed(button: string) {
		return Boolean(sampledButtons & buttonMask(button));
	},
	input_was_just_pressed(button: string) {
		const mask = buttonMask(button);
		return Boolean((sampledButtons & mask) && !(previousButtons & mask));
	},
	input_was_just_released(button: string) {
		const mask = buttonMask(button);
		return Boolean(!(sampledButtons & mask) && (previousButtons & mask));
	},
	write_file(path: string, content: string) {
		micropython.FS.writeFile(projectPath(path), content);
	},
	append_file(path: string, content: string) {
		const target = projectPath(path);
		let previous = '';
		try { previous = micropython.FS.readFile(target, { encoding: 'utf8' }); } catch { /* new file */ }
		micropython.FS.writeFile(target, previous + content);
	},
	write_file_b64(path: string, content: string) {
		micropython.FS.writeFile(projectPath(path), Uint8Array.from(atob(content), (char) => char.charCodeAt(0)));
	},
	append_file_b64(path: string, content: string) {
		const target = projectPath(path);
		let previous = new Uint8Array();
		try { previous = micropython.FS.readFile(target); } catch { /* new file */ }
		const addition = Uint8Array.from(atob(content), (char) => char.charCodeAt(0));
		const combined = new Uint8Array(previous.length + addition.length);
		combined.set(previous); combined.set(addition, previous.length);
		micropython.FS.writeFile(target, combined);
	},
	file_manifest() {
		return micropython.FS.readdir('/project').filter((name: string) => name.endsWith('.py')).sort().join('\n') + '\n';
	},
	delete_file(path: string) {
		try { micropython.FS.unlink(projectPath(path)); } catch { /* matches idempotent editor deletion */ }
	},
	write_image(name: string, width: number, height: number, base64: string) {
		imageUpload = { name, width, height, base64 };
	},
	append_image(name: string, base64: string) {
		if (!imageUpload || imageUpload.name !== name) throw new Error(`no image upload: ${name}`);
		imageUpload.base64 += base64;
	},
	write_image_end(name: string) {
		if (!imageUpload || imageUpload.name !== name) throw new Error(`no image upload: ${name}`);
		images.set(name, decodeBase64Image(name, imageUpload.width, imageUpload.height, imageUpload.base64));
		imageUpload = null;
	},
	image_manifest() {
		return [...images.entries()].map(([name, image]) => `${name},${image.width},${image.height},${checksum(image.base64)}`).join('\n');
	},
	delete_image(name: string) {
		images.delete(name);
	}
} satisfies Record<FerretApiName, (...args: any[]) => any>;

function safePath(name: string) {
	if (!/^[A-Za-z_][A-Za-z0-9_]*\.py$/.test(name)) throw new Error(`Invalid project file: ${name}`);
	return `/project/${name}`;
}

function installProject(project: ProjectData) {
	try { micropython.FS.mkdir('/project'); } catch { /* already exists after a rerun */ }
	for (const oldName of micropython.FS.readdir('/project')) {
		if (oldName !== '.' && oldName !== '..') micropython.FS.unlink(`/project/${oldName}`);
	}
	for (const file of project.codeFiles) micropython.FS.writeFile(safePath(file.name), file.content);
	loadImages(project.images);
}

function pythonBootstrap() {
	return `
import sys
sys.path.insert(0, "/project")
_main_globals = {"__name__": "__main__", "__file__": "/project/main.py"}
with open("/project/main.py", "r") as _main_file:
    _main_source = _main_file.read()
exec(compile(_main_source, "main.py", "exec"), _main_globals, _main_globals)
`;
}

async function initialize() {
	micropython = await loadMicroPython({
		url: micropythonWasmUrl,
		// Match software/cmd/embed/main.go so browser testing exposes the same
		// student-program memory pressure as the RP2350 firmware.
		heapsize: 16 * 1024,
		stdout: (line: string) => emit({ type: 'stdout', text: `${line}\n` }),
		stderr: (line: string) => emit({ type: 'stdout', text: `${line}\n` })
	});
	micropython.registerJsModule('ferret', ferret);
	emit({ type: 'ready' });
}

ctx.onmessage = async (event: MessageEvent<EmulatorCommand>) => {
	const command = event.data;
	if (command.type === 'start') {
		physicalButtons = new Int32Array(command.buttons);
		sampledButtons = previousButtons = 0;
		framebuffer.fill(0);
		try {
			installProject(command.project);
			emit({ type: 'stdout', text: 'Project loaded; executing main.py…\n' });
			micropython.runPython(pythonBootstrap());
			emit({ type: 'finished' });
		} catch (error) {
			emit({ type: 'error', message: error instanceof Error ? error.message : String(error) });
		}
	} else if (command.type === 'repl') {
		try { micropython.runPython(command.text); } catch (error) {
			emit({ type: 'error', message: error instanceof Error ? error.message : String(error) });
		}
	}
};

void initialize().catch((error) => emit({ type: 'error', message: String(error) }));
