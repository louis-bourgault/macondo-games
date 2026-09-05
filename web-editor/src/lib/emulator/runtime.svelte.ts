import type { ProjectData } from '$lib/types';
import { BUTTONS, buttonIndex, type EmulatorButton, type EmulatorCommand, type EmulatorEvent } from './protocol';

export class EmulatorRuntime {
	public ready = $state(false);
	public running = $state(false);
	public output = $state('');
	public frame = $state<Uint16Array | null>(null);
	private worker: Worker | null = null;
	private buttons: SharedArrayBuffer | null = null;
	private loadingTimer: ReturnType<typeof setTimeout> | null = null;

	constructor() {
		this.createWorker();
	}

	private createWorker() {
		this.ready = false;
		if (!crossOriginIsolated) {
			this.output = 'Emulator unavailable: this deployment is missing the required COOP/COEP headers.\n';
		}
		this.worker = new Worker(new URL('./emulator.worker.ts', import.meta.url), { type: 'module' });
		this.loadingTimer = setTimeout(() => {
			if (!this.ready) {
				this.output += 'MicroPython did not load. Check that the worker and .wasm asset are served successfully.\n';
			}
		}, 15000);
		this.worker.onmessage = (event: MessageEvent<EmulatorEvent>) => {
			const message = event.data;
			if (message.type === 'ready') {
				this.ready = true;
				if (this.loadingTimer) clearTimeout(this.loadingTimer);
				this.loadingTimer = null;
			}
			else if (message.type === 'stdout') this.output += message.text;
			else if (message.type === 'frame') this.frame = new Uint16Array(message.pixels);
			else if (message.type === 'finished') this.running = false;
			else if (message.type === 'error') {
				this.output += `${message.message}\n`;
				this.running = false;
			}
		};
		this.worker.onerror = (event) => {
			const location = event.filename ? ` (${event.filename}:${event.lineno}:${event.colno})` : '';
			this.output += `Emulator worker failed to load: ${event.message || 'unknown module/WASM loading error'}${location}\n`;
			this.running = false;
			this.ready = false;
			event.preventDefault();
		};
	}

	public start(project: ProjectData) {
		if (!this.worker || !this.ready) throw new Error('The emulator is still loading.');
		if (this.running) this.stop();
		if (!crossOriginIsolated) throw new Error('The emulator requires cross-origin isolation for live button input.');
		this.buttons = new SharedArrayBuffer(BUTTONS.length * 4);
		this.output = 'Starting Ferretboard RP2350 emulator…\n';
		// Svelte's deeply reactive project is a Proxy, which the worker structured
		// clone algorithm cannot transfer. JSON also mirrors project downloads.
		const snapshot = JSON.parse(JSON.stringify(project)) as ProjectData;
		this.worker?.postMessage({ type: 'start', project: snapshot, buttons: this.buttons } satisfies EmulatorCommand);
		this.running = true;
	}

	public stop() {
		this.worker?.terminate();
		this.worker = null;
		this.running = false;
		this.buttons = null;
		this.output += '--- Emulator stopped ---\n';
		this.createWorker();
	}

	public setButton(button: EmulatorButton, pressed: boolean) {
		if (!this.buttons) return;
		Atomics.store(new Int32Array(this.buttons), buttonIndex(button), pressed ? 1 : 0);
	}

	public sendTerminalInput(text: string) {
		if (this.running) {
			this.output += 'Stop the running game before using the emulator REPL.\n';
			return;
		}
		this.worker?.postMessage({ type: 'repl', text } satisfies EmulatorCommand);
	}

	public destroy() {
		if (this.loadingTimer) clearTimeout(this.loadingTimer);
		this.worker?.terminate();
		this.worker = null;
	}
}
