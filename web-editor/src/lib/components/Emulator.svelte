<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import type { ProjectData } from '$lib/types';
	import { EmulatorRuntime } from '$lib/emulator/runtime.svelte';
	import type { EmulatorButton } from '$lib/emulator/protocol';

	let { project }: { project: ProjectData } = $props();
	let canvas: HTMLCanvasElement | null = $state(null);
	const runtime = new EmulatorRuntime();
	const rgba = new Uint8ClampedArray(240 * 240 * 4);

	const keyMap: Record<string, EmulatorButton> = {
		ArrowUp: 'UP', w: 'UP', W: 'UP', ArrowDown: 'DOWN', s: 'DOWN', S: 'DOWN',
		ArrowLeft: 'LEFT', a: 'LEFT', ArrowRight: 'RIGHT', d: 'RIGHT', D: 'RIGHT',
		j: 'A', J: 'A', k: 'B', K: 'B', u: 'X', U: 'X', i: 'Y', I: 'Y', Enter: 'MENU'
	};

	$effect(() => {
		const frame = runtime.frame;
		if (!canvas || !frame) return;
		for (let i = 0; i < frame.length; i++) {
			const color = frame[i];
			const r = (color >> 11) & 0x1f;
			const g = (color >> 5) & 0x3f;
			const b = color & 0x1f;
			const offset = i * 4;
			rgba[offset] = (r << 3) | (r >> 2);
			rgba[offset + 1] = (g << 2) | (g >> 4);
			rgba[offset + 2] = (b << 3) | (b >> 2);
			rgba[offset + 3] = 255;
		}
		canvas.getContext('2d')?.putImageData(new ImageData(rgba, 240, 240), 0, 0);
	});

	function keyboard(event: KeyboardEvent, pressed: boolean) {
		const button = keyMap[event.key];
		if (!button) return;
		event.preventDefault();
		runtime.setButton(button, pressed);
	}

	function releaseAll() {
		for (const button of ['UP', 'DOWN', 'LEFT', 'RIGHT', 'A', 'B', 'X', 'Y', 'MENU'] as EmulatorButton[]) {
			runtime.setButton(button, false);
		}
	}

	function pointer(button: EmulatorButton, pressed: boolean) {
		runtime.setButton(button, pressed);
	}

	onDestroy(() => runtime.destroy());
</script>

<section class="flex h-full min-h-0 flex-col gap-3 overflow-y-auto p-3" aria-label="Ferretboard emulator">
	<div class="flex flex-wrap items-center gap-2">
		<Button onclick={() => runtime.start(project)} disabled={!runtime.ready || runtime.running}>
			{runtime.ready ? (runtime.running ? 'Running…' : 'Run in emulator') : 'Loading MicroPython…'}
		</Button>
		<Button variant="secondary" onclick={() => runtime.stop()} disabled={!runtime.running}>Stop</Button>
	</div>

	<dialog
		open
		class="mx-auto rounded-3xl border bg-zinc-900 p-4 shadow-lg outline-none focus:ring-2 focus:ring-ring"
		tabindex="0"
		onkeydown={(event) => keyboard(event, true)}
		onkeyup={(event) => keyboard(event, false)}
		onblur={releaseAll}
	>
		<canvas bind:this={canvas} width="240" height="240" class="block aspect-square w-full max-w-80 bg-black [image-rendering:pixelated]"></canvas>
		<div class="mt-4 grid grid-cols-[1fr_auto_1fr] items-center gap-5 select-none">
			<div class="grid grid-cols-3 grid-rows-3 gap-1 justify-self-center">
				<span></span>{@render Control('▲', 'UP', pointer)}<span></span>
				{@render Control('◀', 'LEFT', pointer)}<span></span>{@render Control('▶', 'RIGHT', pointer)}
				<span></span>{@render Control('▼', 'DOWN', pointer)}<span></span>
			</div>
			{@render Control('MENU', 'MENU', pointer, true)}
			<div class="grid grid-cols-3 grid-rows-3 gap-1 justify-self-center">
				<span></span>{@render Control('Y', 'Y', pointer)}<span></span>
				{@render Control('X', 'X', pointer)}<span></span>{@render Control('B', 'B', pointer)}
				<span></span>{@render Control('A', 'A', pointer)}<span></span>
			</div>
		</div>
	</dialog>

	<p class="text-xs text-muted-foreground">Click the console first. Arrows/WASD: d-pad · J/K: A/B · U/I: X/Y · Enter: MENU</p>
	<pre class="min-h-24 flex-1 overflow-auto rounded border bg-muted p-2 text-xs whitespace-pre-wrap">{runtime.output || 'MicroPython output will appear here.'}</pre>
</section>

{#snippet Control(label: string, button: EmulatorButton, pointer: (button: EmulatorButton, pressed: boolean) => void, small = false)}
	<button
		type="button"
		class={small ? 'h-7 rounded-full bg-zinc-600 px-3 text-[10px] text-white active:bg-zinc-400' : 'size-10 rounded-full bg-zinc-700 font-bold text-white active:bg-zinc-400'}
		onpointerdown={(event) => { event.currentTarget.setPointerCapture(event.pointerId); pointer(button, true); }}
		onpointerup={() => pointer(button, false)}
		onpointercancel={() => pointer(button, false)}
	>{label}</button>
{/snippet}
