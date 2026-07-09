<script lang="ts">
	import Button from './ui/button/button.svelte';
	import {
		Save,
		Upload,
		Pencil,
		Eraser,
		Pipette,
		Trash,
		Undo,
		Redo,
		Plus,
		Minus
	} from '@lucide/svelte';
	import { type ImageFile } from '$lib/types';
	import { onMount, untrack } from 'svelte';

	let { imageData = $bindable() }: { imageData: ImageFile } = $props();

	type Tool = 'brush' | 'eraser' | 'eyedropper';

	// rgb565 chroma key transparent
	const MAGENTA_KEY = 0xf81f;
	const MAGENTA_NUDGE = 0xf83f; // opaque magenta bumped to g=1 so it survives a round-trip

	const swatches: { transparent?: boolean; colour?: string }[] = [
		{ transparent: true },
		{ colour: '#000000' },
		{ colour: '#ffffff' },
		{ colour: '#ff0000' },
		{ colour: '#ffa500' },
		{ colour: '#ffff00' },
		{ colour: '#00ff00' },
		{ colour: '#0000ff' }
	];

	let brushSize = $state(2);
	let selectedColour = $state('#000000');
	let transparentSelected = $state(false);
	let tool = $state<Tool>('brush');

	let canvas: HTMLCanvasElement | null = $state(null);
	let scrollEl: HTMLDivElement | null = $state(null);
	let fileInput: HTMLInputElement | null = $state(null);
	let ctx: CanvasRenderingContext2D | null = null;

	let isDrawing = false;
	let lastX = 0;
	let lastY = 0;

	let areaW = $state(0);
	let areaH = $state(0);
	let scale = $state(0);
	let userZoomed = $state(false);
	let undoStack = $state<ImageData[]>([]);
	let redoStack = $state<ImageData[]>([]);
	const MAX_UNDO = 40;
	const MAX_SCALE = 32;

	let fitScale = $derived(
		areaW && areaH && imageData.width && imageData.height
			? Math.max(1, Math.floor(Math.min(areaW / imageData.width, areaH / imageData.height)))
			: 1
	);
	let displayScale = $derived(userZoomed ? scale : fitScale);
	let paintingTransparent = $derived(
		tool === 'eraser' || (tool === 'brush' && transparentSelected)
	);

	function clamp(v: number, lo: number, hi: number) {
		return v < lo ? lo : v > hi ? hi : v;
	}

	function setBrushSize(v: number) {
		brushSize = clamp(v, 1, 10);
	}

	function zoomIn() {
		scale = clamp((userZoomed ? scale : fitScale) + 1, 1, MAX_SCALE);
		userZoomed = true;
	}
	function zoomOut() {
		scale = clamp((userZoomed ? scale : fitScale) - 1, 1, MAX_SCALE);
		userZoomed = true;
	}
	function resetFit() {
		userZoomed = false;
	}

	function clearToTransparent(c: CanvasRenderingContext2D) {
		c.clearRect(0, 0, imageData.width, imageData.height);
	}

	function loadImage(c: CanvasRenderingContext2D, content: string) {
		const w = imageData.width;
		const h = imageData.height;
		if (!w || !h || !canvas) return;
		canvas.width = w;
		canvas.height = h;
		c.imageSmoothingEnabled = false;
		c.globalCompositeOperation = 'source-over';
		if (content) {
			try {
				const bytes = bytesFromBase64(content);
				if (bytes.length !== w * h * 2) {
					clearToTransparent(c);
					return;
				}
				const img = c.createImageData(w, h);
				const d = img.data;
				for (let i = 0, o = 0; i < d.length; i += 4, o += 2) {
					// little-endian RGB565 (matches MicroPython framebuf on RP2040)
					const v = bytes[o] | (bytes[o + 1] << 8);
					if (v === MAGENTA_KEY) {
						// chroma key → transparent
						d[i] = 0;
						d[i + 1] = 0;
						d[i + 2] = 0;
						d[i + 3] = 0;
					} else {
						const r5 = (v >> 11) & 0x1f;
						const g6 = (v >> 5) & 0x3f;
						const b5 = v & 0x1f;
						d[i] = (r5 << 3) | (r5 >> 2);
						d[i + 1] = (g6 << 2) | (g6 >> 4);
						d[i + 2] = (b5 << 3) | (b5 >> 2);
						d[i + 3] = 255;
					}
				}
				c.putImageData(img, 0, 0);
			} catch {
				clearToTransparent(c);
			}
		} else {
			clearToTransparent(c);
		}
		undoStack = [];
		redoStack = [];
	}

	$effect(() => {
		const _name = imageData.name;
		const _w = imageData.width;
		const _h = imageData.height;
		if (!canvas) return;
		const c = canvas.getContext('2d');
		if (!c) return;
		ctx = c;
		const content = untrack(() => imageData.content);
		loadImage(c, content);
		userZoomed = false;
	});

	onMount(() => {
		const ro = new ResizeObserver((entries) => {
			const cr = entries[0]?.contentRect;
			if (!cr) return;
			areaW = cr.width;
			areaH = cr.height;
		});
		if (scrollEl) ro.observe(scrollEl);
		return () => ro.disconnect();
	});

	function getPixel(e: PointerEvent) {
		if (!canvas) return { x: 0, y: 0 };
		const rect = canvas.getBoundingClientRect();
		const x = Math.floor((e.clientX - rect.left) / displayScale);
		const y = Math.floor((e.clientY - rect.top) / displayScale);
		return { x: clamp(x, 0, imageData.width - 1), y: clamp(y, 0, imageData.height - 1) };
	}

	function stamp(x: number, y: number) {
		if (!ctx) return;
		ctx.fillStyle = paintingTransparent ? '#000' : selectedColour;
		const half = Math.floor(brushSize / 2);
		ctx.fillRect(x - half, y - half, brushSize, brushSize);
	}

	function drawLine(x0: number, y0: number, x1: number, y1: number) {
		let dx = Math.abs(x1 - x0);
		let dy = -Math.abs(y1 - y0);
		const sx = x0 < x1 ? 1 : -1;
		const sy = y0 < y1 ? 1 : -1;
		let err = dx + dy;
		for (;;) {
			stamp(x0, y0);
			if (x0 === x1 && y0 === y1) break;
			const e2 = 2 * err;
			if (e2 >= dy) {
				err += dy;
				x0 += sx;
			}
			if (e2 <= dx) {
				err += dx;
				y0 += sy;
			}
		}
	}

	function pickColor(x: number, y: number) {
		if (!ctx) return;
		const d = ctx.getImageData(x, y, 1, 1).data;
		if (d[3] === 0) {
			transparentSelected = true;
		} else {
			transparentSelected = false;
			selectedColour =
				'#' + [d[0], d[1], d[2]].map((v) => v.toString(16).padStart(2, '0')).join('');
		}
		tool = 'brush';
	}

	function pushUndo() {
		if (!ctx) return;
		undoStack.push(ctx.getImageData(0, 0, imageData.width, imageData.height));
		if (undoStack.length > MAX_UNDO) undoStack.shift();
		redoStack = [];
	}

	function undo() {
		if (!ctx || undoStack.length === 0) return;
		redoStack.push(ctx.getImageData(0, 0, imageData.width, imageData.height));
		const snap = undoStack.pop();
		if (snap) ctx.putImageData(snap, 0, 0);
	}

	function redo() {
		if (!ctx || redoStack.length === 0) return;
		undoStack.push(ctx.getImageData(0, 0, imageData.width, imageData.height));
		const snap = redoStack.pop();
		if (snap) ctx.putImageData(snap, 0, 0);
	}

	function onPointerDown(e: PointerEvent) {
		if (!ctx) return;
		e.preventDefault();
		canvas?.setPointerCapture(e.pointerId);
		const { x, y } = getPixel(e);
		if (tool === 'eyedropper') {
			pickColor(x, y);
			return;
		}
		pushUndo();
		isDrawing = true;
		lastX = x;
		lastY = y;
		ctx.globalCompositeOperation = paintingTransparent ? 'destination-out' : 'source-over';
		stamp(x, y);
	}

	function onPointerMove(e: PointerEvent) {
		if (!isDrawing) return;
		const { x, y } = getPixel(e);
		drawLine(lastX, lastY, x, y);
		lastX = x;
		lastY = y;
	}

	function onPointerUp(e: PointerEvent) {
		isDrawing = false;
		if (ctx) ctx.globalCompositeOperation = 'source-over';
		try {
			canvas?.releasePointerCapture(e.pointerId);
		} catch {
			// ignore
		}
	}

	function clearCanvas() {
		if (!ctx) return;
		pushUndo();
		ctx.globalCompositeOperation = 'source-over';
		clearToTransparent(ctx);
	}

	function saveData() {
		if (!ctx) return;
		const img = ctx.getImageData(0, 0, imageData.width, imageData.height);
		imageData.content = rgb565ToBase64(img);
	}

	function uploadImage() {
		fileInput?.click();
	}

	function onFileSelected(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		const url = URL.createObjectURL(file);
		const img = new Image();
		img.onload = () => {
			if (!ctx) return;
			pushUndo();
			ctx.globalCompositeOperation = 'source-over';
			clearToTransparent(ctx);
			ctx.imageSmoothingEnabled = true;
			ctx.drawImage(img, 0, 0, imageData.width, imageData.height);
			ctx.imageSmoothingEnabled = false;
			// binarise alpha: no half-transparency allowed
			const id = ctx.getImageData(0, 0, imageData.width, imageData.height);
			const d = id.data;
			for (let i = 0; i < d.length; i += 4) {
				if (d[i + 3] < 128) {
					d[i] = 0;
					d[i + 1] = 0;
					d[i + 2] = 0;
					d[i + 3] = 0;
				} else {
					d[i + 3] = 255;
				}
			}
			ctx.putImageData(id, 0, 0);
			URL.revokeObjectURL(url);
			input.value = '';
		};
		img.onerror = () => {
			URL.revokeObjectURL(url);
			input.value = '';
		};
		img.src = url;
	}

	function rgb565ToBase64(img: ImageData): string {
		const { width, height, data } = img;
		const out = new Uint8Array(width * height * 2);
		let o = 0;
		for (let i = 0; i < data.length; i += 4) {
			let v: number;
			if (data[i + 3] < 128) {
				// transparent → chroma key
				v = MAGENTA_KEY;
			} else {
				const r = data[i] >> 3;
				const g = data[i + 1] >> 2;
				const b = data[i + 2] >> 3;
				v = (r << 11) | (g << 5) | b;
				if (v === MAGENTA_KEY) v = MAGENTA_NUDGE; // keep opaque magenta visible
			}
			out[o++] = v & 0xff;
			out[o++] = (v >> 8) & 0xff;
		}
		return base64FromBytes(out);
	}

	function base64FromBytes(bytes: Uint8Array): string {
		let bin = '';
		const chunk = 0x8000;
		for (let i = 0; i < bytes.length; i += chunk) {
			bin += String.fromCharCode(...bytes.subarray(i, i + chunk));
		}
		return btoa(bin);
	}

	function bytesFromBase64(b64: string): Uint8Array {
		const bin = atob(b64);
		const out = new Uint8Array(bin.length);
		for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
		return out;
	}

	function swatchActive(sw: { transparent?: boolean; colour?: string }) {
		return sw.transparent
			? transparentSelected
			: !transparentSelected &&
					sw.colour !== undefined &&
					selectedColour.toLowerCase() === sw.colour;
	}

	function selectSwatch(sw: { transparent?: boolean; colour?: string }) {
		if (sw.transparent) {
			transparentSelected = true;
		} else if (sw.colour) {
			transparentSelected = false;
			selectedColour = sw.colour;
		}
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex flex-col gap-2 border-b border-border p-2">
		<div class="flex items-center justify-between gap-2">
			<span class="truncate text-sm font-medium">{imageData.name}</span>
			<span class="shrink-0 text-xs text-muted-foreground">
				{imageData.width}×{imageData.height} · {displayScale}×
			</span>
		</div>

		<div class="flex flex-wrap items-center gap-1">
			<Button
				size="icon-sm"
				variant="outline"
				onclick={undo}
				disabled={undoStack.length === 0}
				title="Undo"
			>
				<Undo />
			</Button>
			<Button
				size="icon-sm"
				variant="outline"
				onclick={redo}
				disabled={redoStack.length === 0}
				title="Redo"
			>
				<Redo />
			</Button>
			<span class="mx-1 w-px self-stretch bg-border"></span>
			<span class="text-xs text-muted-foreground">Zoom</span>
			<Button size="icon-sm" variant="outline" onclick={zoomOut}>
				<Minus />
			</Button>
			<span class="w-6 text-center text-xs tabular-nums">{displayScale}×</span>
			<Button size="icon-sm" variant="outline" onclick={zoomIn}>
				<Plus />
			</Button>
			<Button size="xs" variant="outline" onclick={resetFit}>Fit</Button>
			<span class="mx-1 w-px self-stretch bg-border"></span>
			<Button size="icon-sm" variant="outline" onclick={clearCanvas} title="Clear">
				<Trash />
			</Button>
			<span class="mx-1 w-px self-stretch bg-border"></span>
			<Button size="icon-sm" onclick={saveData} title="Save">
				<Save />
			</Button>
			<Button size="icon-sm" variant="outline" onclick={uploadImage} title="Upload image">
				<Upload />
			</Button>
		</div>

		<div class="flex flex-wrap items-center gap-1">
			<Button
				size="icon-sm"
				variant={tool === 'brush' ? 'default' : 'outline'}
				onclick={() => (tool = 'brush')}
				title="Brush"
			>
				<Pencil />
			</Button>
			<Button
				size="icon-sm"
				variant={tool === 'eraser' ? 'default' : 'outline'}
				onclick={() => (tool = 'eraser')}
				title="Eraser"
			>
				<Eraser />
			</Button>
			<Button
				size="icon-sm"
				variant={tool === 'eyedropper' ? 'default' : 'outline'}
				onclick={() => (tool = 'eyedropper')}
				title="Eyedropper"
			>
				<Pipette />
			</Button>
			<span class="mx-1 w-px self-stretch bg-border"></span>
			<span class="text-xs text-muted-foreground">Brush</span>
			<Button size="icon-sm" variant="outline" onclick={() => setBrushSize(brushSize - 1)}>
				<Minus />
			</Button>
			<span class="w-5 text-center text-xs tabular-nums">{brushSize}</span>
			<Button size="icon-sm" variant="outline" onclick={() => setBrushSize(brushSize + 1)}>
				<Plus />
			</Button>
			<span class="mx-1 w-px self-stretch bg-border"></span>
			<input
				type="color"
				bind:value={selectedColour}
				onchange={() => (transparentSelected = false)}
				title="Custom colour"
				class="size-7 cursor-pointer border border-border bg-transparent p-0.5"
			/>
			<span class="mx-1 w-px self-stretch bg-border"></span>
			<div class="flex items-center gap-1">
				{#each swatches as sw (sw.transparent ? 't' : sw.colour)}
					<button
						type="button"
						onclick={() => selectSwatch(sw)}
						title={sw.transparent ? 'Transparent' : sw.colour}
						aria-label={sw.transparent ? 'Transparent' : sw.colour}
						class="size-6 shrink-0 border border-border p-0 {sw.transparent
							? 'checkerboard'
							: ''} {swatchActive(sw) ? 'ring-2 ring-ring' : ''}"
						style={sw.transparent ? '' : `background-color:${sw.colour}`}
					></button>
				{/each}
			</div>
		</div>
	</div>

	<div bind:this={scrollEl} class="canvas-area min-h-0 flex-1 overflow-auto bg-muted/30">
		<div class="flex min-h-full min-w-full items-center justify-center p-4">
			<div
				class="relative shrink-0 border border-border"
				style="width:{imageData.width * displayScale}px;height:{imageData.height * displayScale}px"
			>
				<canvas
					bind:this={canvas}
					width={imageData.width}
					height={imageData.height}
					onpointerdown={onPointerDown}
					onpointermove={onPointerMove}
					onpointerup={onPointerUp}
					onpointercancel={onPointerUp}
					class="checkerboard block touch-none select-none"
					style="width:{imageData.width * displayScale}px;height:{imageData.height *
						displayScale}px;image-rendering:pixelated;cursor:crosshair"
				></canvas>
			</div>
		</div>
	</div>

	<input
		type="file"
		accept="image/*"
		bind:this={fileInput}
		onchange={onFileSelected}
		class="hidden"
	/>
</div>

<style>
	.checkerboard {
		background-color: #ffffff;
		background-image:
			linear-gradient(45deg, #d4d4d4 25%, transparent 25%, transparent 75%, #d4d4d4 75%),
			linear-gradient(45deg, #d4d4d4 25%, transparent 25%, transparent 75%, #d4d4d4 75%);
		background-size: 20px 20px;
		background-position:
			0 0,
			10px 10px;
	}
</style>
