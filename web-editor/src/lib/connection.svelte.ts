import type { ProjectData, ImageFile } from './types';

const MAX_IMAGE_WIDTH = 240;
const MAX_IMAGE_HEIGHT = 240;
const MAX_CODE_FILE_BYTES = 16 * 1024;
const CODE_FILE_NAME = /^[A-Za-z_][A-Za-z0-9_]*\.py$/;

export function isValidCodeFileName(name: string): boolean {
	return CODE_FILE_NAME.test(name);
}

export class WebSerialConnection {
	public port: any;
	public reader: any;
	public writer: any;
	public keepreading = false;
	public readableStreamClosed: any;
	public writableStreamClosed: any;
	public connected = $state(false);
	public output = $state('');
	private attached = false;

	private delay(ms: number) {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	private async waitForOutputSince(
		needle: string,
		startLen: number,
		timeoutMs = 5000
	): Promise<void> {
		const deadline = Date.now() + timeoutMs;
		while (Date.now() < deadline) {
			if (this.output.slice(startLen).includes(needle)) return;
			await this.delay(10);
		}
		throw new Error(`timeout waiting for ${needle}`);
	}

	private async pasteAndRun(source: string): Promise<string> {
		// Do not rely on a fixed delay here. A payload sent before Ctrl-E has
		// actually reached the device is interpreted by the friendly REPL.
		const pasteStart = this.output.length;
		await this.controlE();
		await this.waitForOutputSince('=== ', pasteStart);

		// Take the checkpoint before Ctrl-D: a quick device can print the prompt
		// before the next await gets to inspect output.
		const responseStart = this.output.length;
		await this.sendText(source);
		await this.controlD();
		await this.waitForOutputSince('>>> ', responseStart);
		const response = this.output.slice(responseStart);
		if (response.includes('Traceback (most recent call last):')) {
			throw new Error(`device rejected paste:\n${response}`);
		}
		return response;
	}

	private async attachPort(port: any) {
		if (this.attached && this.port === port) {
			this.connected = true;
			return;
		}

		this.port = port;
		if (this.port.readable?.locked || this.port.writable?.locked) {
			this.log('\n--- Port already in use ---\n');
			return;
		}
		const textDecoder = new TextDecoderStream();
		this.readableStreamClosed = this.port.readable.pipeTo(textDecoder.writable);
		this.reader = textDecoder.readable.getReader();

		const textEncoder = new TextEncoderStream();
		this.writableStreamClosed = textEncoder.readable.pipeTo(this.port.writable);
		this.writer = textEncoder.writable.getWriter();

		this.keepreading = true;
		void this.readLoop();
		this.connected = true;
		this.attached = true;
	}

	// FNV-1a 32-bit checksum, simple and fast, not cryptographically secure
	private generateChecksum(str: string): string {
		let hash = 0x811c9dc5;

		for (let i = 0; i < str.length; i++) {
			hash ^= str.charCodeAt(i);
			hash = Math.imul(hash, 0x01000193);
		}

		return (hash >>> 0).toString(16).padStart(8, '0');
	}

	private imageNameLiteral(name: string): string {
		// JSON string escaping is also valid Python string-literal escaping for the
		// names accepted by the firmware.
		return JSON.stringify(name);
	}

	private validateImages(imgs: ImageFile[]) {
		const names = new Set<string>();
		for (const img of imgs) {
			if (!img.name || /[\\/:,\r\n]/.test(img.name) || img.name === '.' || img.name === '..') {
				throw new Error(`Image name "${img.name}" is invalid.`);
			}
			if (names.has(img.name)) {
				throw new Error(`Image name "${img.name}" is duplicated.`);
			}
			names.add(img.name);
			if (
				!Number.isInteger(img.width) ||
				!Number.isInteger(img.height) ||
				img.width < 1 ||
				img.height < 1 ||
				img.width > MAX_IMAGE_WIDTH ||
				img.height > MAX_IMAGE_HEIGHT
			) {
				throw new Error(
					`Image "${img.name}" must have integer dimensions from 1×1 to ${MAX_IMAGE_WIDTH}×${MAX_IMAGE_HEIGHT}.`
				);
			}

			const expectedBytes = img.width * img.height * 2;
			const expectedBase64Length = 4 * Math.ceil(expectedBytes / 3);
			if (typeof img.content !== 'string' || img.content.length !== expectedBase64Length) {
				throw new Error(`Image "${img.name}" data does not match its dimensions.`);
			}
			try {
				if (atob(img.content).length !== expectedBytes) {
					throw new Error('decoded size mismatch');
				}
			} catch {
				throw new Error(`Image "${img.name}" contains invalid base64 data.`);
			}
		}
	}

	private validateCodeFiles(projectData: ProjectData) {
		const names = new Set<string>();
		for (const file of projectData.codeFiles) {
			if (!isValidCodeFileName(file.name)) {
				throw new Error(
					`Code file name "${file.name}" is invalid. Use a Python identifier ending in .py.`
				);
			}
			if (names.has(file.name)) {
				throw new Error(`Code file name "${file.name}" is duplicated.`);
			}
			names.add(file.name);
			if (new TextEncoder().encode(file.content).length > MAX_CODE_FILE_BYTES) {
				throw new Error(`Code file "${file.name}" exceeds the 16 KiB device limit.`);
			}
		}
		if (!names.has('main.py')) {
			throw new Error('The project must contain main.py.');
		}
	}

	private bytesToBase64(bytes: Uint8Array): string {
		let binary = '';
		for (const byte of bytes) binary += String.fromCharCode(byte);
		return btoa(binary);
	}

	private async getDeviceCodeFiles(): Promise<Set<string>> {
		const begin = '__FERRET_CODE_FILES_BEGIN__';
		const end = '__FERRET_CODE_FILES_END__';
		const response = await this.pasteAndRun(
			`import ferret\nprint("${begin}")\nprint(ferret.file_manifest(), end="")\nprint("${end}")\n`
		);
		// Paste mode echoes the source, so use the final marker pair emitted by
		// the executed program rather than the same text in the echoed source.
		const start = response.lastIndexOf(begin);
		const finish = response.lastIndexOf(end);
		if (start < 0 || finish < start) {
			throw new Error('device returned an unreadable code-file manifest');
		}
		return new Set(
			response
				.slice(start + begin.length, finish)
				.split(/\r?\n/)
				.map((name) => name.trim())
				.filter(Boolean)
		);
	}

	private async syncCodeFiles(projectData: ProjectData) {
		this.validateCodeFiles(projectData);
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		const existingFiles = await this.getDeviceCodeFiles();

		for (const file of projectData.codeFiles) {
			await this.writeFileToDevice(file.name, file.content);
		}

		const wantedFiles = new Set(projectData.codeFiles.map((file) => file.name));
		for (const filename of existingFiles) {
			if (!wantedFiles.has(filename)) {
				await this.pasteAndRun(`import ferret\nferret.delete_file(${JSON.stringify(filename)})\n`);
			}
		}
	}

	public async syncImages(imgs: ImageFile[]) {
		console.log('received images to sync', imgs);
		this.validateImages(imgs);

		// Stop whatever is running so the REPL is responsive, then read the
		// image manifest that the firmware keeps on its own filesystem.
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		// Note where output ends before querying the manifest so we only parse
		// the response to THIS query. connection.output accumulates the whole
		// session (including the controlC prompts above), so splitting on the
		// first '>>> ' would grab the stale bytes between the first two prompts.
		const manifestStart = this.output.length;
		await this.pasteAndRun(`import ferret\nprint(ferret.image_manifest())\n`);

		const existingImagesData = this.output.slice(manifestStart);
		console.log('existing images data', existingImagesData);

		const existingImagesMap = new Map<
			string,
			{ width: number; height: number; checksum: string }
		>();
		for (const rawLine of existingImagesData.split('\n')) {
			// trim so a trailing \r (the device sends CRLF) doesn't corrupt the checksum
			const [name, widthStr, heightStr, checksum] = rawLine.trim().split(',');
			if (name && widthStr && heightStr && checksum && !isNaN(parseInt(widthStr, 10))) {
				existingImagesMap.set(name, {
					width: parseInt(widthStr, 10),
					height: parseInt(heightStr, 10),
					checksum
				});
			}
		}

		// TinyUSB has a 512-byte receive staging buffer. Keep the complete Python
		// paste (wrapper plus payload) comfortably below it; the previous 2 KiB
		// writes could overrun that buffer before the RP2040 drained it.
		const chunkSize = 320;

		for (const img of imgs) {
			const checksum = this.generateChecksum(img.content);
			if (existingImagesMap.has(img.name)) {
				const existing = existingImagesMap.get(img.name);
				if (
					existing &&
					existing.width === img.width &&
					existing.height === img.height &&
					existing.checksum === checksum
				) {
					console.log(`Skipping ${img.name}, hasn't chnaged since last checksum`);
					continue;
				}
			}

			// Upload in chunks: write_image opens a temporary file, append_image
			// streams more base64 into it, and write_image_end validates and publishes
			// the finished image plus its manifest entry.
			const nameLiteral = this.imageNameLiteral(img.name);
			const totalChunks = Math.ceil(img.content.length / chunkSize);
			let completedChunks = 0;
			try {
				for (let i = 0; i < img.content.length; i += chunkSize) {
					const chunk = img.content.slice(i, i + chunkSize);
					const call =
						i === 0
							? `ferret.write_image(${nameLiteral}, ${img.width}, ${img.height}, "${chunk}")`
							: `ferret.append_image(${nameLiteral}, "${chunk}")`;
					await this.pasteAndRun(`import ferret\n${call}\n`);
					completedChunks++;
				}

				await this.pasteAndRun(`import ferret\nferret.write_image_end(${nameLiteral})\n`);
			} catch (err) {
				throw new Error(
					`Failed to upload image "${img.name}" after ${completedChunks}/${totalChunks} chunks: ${err instanceof Error ? err.message : String(err)}`
				);
			}
		}

		for (const name of existingImagesMap.keys()) {
			if (!imgs.find((img) => img.name === name)) {
				//delete the image from the device if it is not in the new list
				const nameLiteral = this.imageNameLiteral(name);
				await this.pasteAndRun(
					`import ferret\ntry:\n    ferret.delete_image(${nameLiteral})\nexcept OSError:\n    pass\n`
				);
			}
		}
	}

	public init = async () => {
		// Don't auto-open any port — the user must click Connect to get the
		// port selector. Auto-opening a previously-granted port would set
		// connected=true and make connect() early-return without showing
		// the selector.
	};

	public connect = async () => {
		const serial = (navigator as Navigator & { serial?: any }).serial;
		if (!serial) {
			alert('Web Serial API not supported. Use Chrome or Edge over HTTPS (or localhost).');
			return;
		}
		if (this.connected && this.port) {
			return;
		}
		try {
			this.port = await serial.requestPort();
			await this.port.open({ baudRate: 115200 });
			await this.attachPort(this.port);
			this.log('Connected to serial port.\n');
		} catch (err: any) {
			if (err.name !== 'NotFoundError') {
				this.log(`\n--- Connection error: ${err.message} ---\n`);
			}
		}
	};

	public controlD = async () => {
		await this.sendText('\x04');
	};

	public controlC = async () => {
		await this.sendText('\x03');
	};

	public controlE = async () => {
		await this.sendText('\x05');
	};

	public controlB = async () => {
		await this.sendText('\x02');
	};

	public sendText = async (text: string) => {
		if (!this.writer) {
			this.log('\n--- Not connected: writer is not available ---\n');
			return;
		}
		await this.writer.write(text);
	};

	public runScript = async (script: string) => {
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		await this.pasteAndRun(
			'import gc\n' +
				'for _n in list(globals()):\n' +
				'    if not _n.startswith("_") and _n != "gc":\n' +
				'        try:\n' +
				'            del globals()[_n]\n' +
				'        except:\n' +
				'            pass\n' +
				'gc.collect()\n' +
				'\n' +
				script +
				'\n'
		); //RUN the garbage colelctor; avoids out of memory issues.
	};

	public saveProjectToDevice = async (projectData: ProjectData) => {
		await this.syncCodeFiles(projectData);
		// Boot main.py right away: a soft reset re-runs the boot script, so the
		// game starts immediately instead of needing a power cycle.
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		await this.controlD();
	};

	public writeFileToDevice = async (filename: string, content: string) => {
		if (!isValidCodeFileName(filename)) {
			throw new Error(`Code file name "${filename}" is invalid.`);
		}
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);

		// Encode UTF-8 bytes rather than interpolating source into a Python
		// literal. This preserves arbitrary quotes, backslashes and Unicode.
		// A zero-byte file still sends one write call, which truncates old data.
		const bytes = new TextEncoder().encode(content);
		if (bytes.length > MAX_CODE_FILE_BYTES) {
			throw new Error(`Code file "${filename}" exceeds the 16 KiB device limit.`);
		}
		const rawChunkSize = 240;
		const chunks = Math.max(1, Math.ceil(bytes.length / rawChunkSize));
		for (let chunkIndex = 0; chunkIndex < chunks; chunkIndex++) {
			const start = chunkIndex * rawChunkSize;
			const chunk = this.bytesToBase64(bytes.slice(start, start + rawChunkSize));
			const call =
				chunkIndex === 0
					? `ferret.write_file_b64(${JSON.stringify(filename)}, "${chunk}")`
					: `ferret.append_file_b64(${JSON.stringify(filename)}, "${chunk}")`;
			await this.pasteAndRun(`import ferret\n${call}\n`);
		}
	};

	public runProject = async (projectData: ProjectData) => {
		// Persist every file (including main.py) so `import` statements resolve
		// and the boot script has fresh content, then soft-reset: main.py runs
		// automatically from flash with a fresh namespace.
		await this.syncCodeFiles(projectData);
		// Stop anything currently running and land at a clean prompt.
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		// Reboot: the firmware boots /main.py on every soft reset.
		await this.controlD();
	};

	public log = (msg: string) => {
		this.output += msg;
		console.log(msg);
	};

	public readLoop = async () => {
		try {
			while (this.keepreading) {
				const { value, done } = await this.reader.read();
				if (done) {
					this.reader.releaseLock();
					break;
				}
				if (value) this.log(value);
			}
		} catch (err: any) {
			this.log(`\n--- Read error: ${err.message} ---\n`);
		}
		this.connected = false;
		this.attached = false;
		this.keepreading = false;
		this.reader = null;
		this.writer = null;
		this.port = null;
	};

	public disconnect = async () => {
		this.keepreading = false;
		try {
			if (this.reader) {
				await this.reader.cancel();
				this.reader = null;
			}
			if (this.writer) {
				await this.writer.close();
				this.writer = null;
			}
			if (this.readableStreamClosed) {
				await this.readableStreamClosed.catch(() => {});
			}
			if (this.writableStreamClosed) {
				await this.writableStreamClosed.catch(() => {});
			}
			if (this.port) {
				await this.port.close();
				this.port = null;
			}
		} catch (err: any) {
			this.log(`\n--- Disconnect error: ${err.message} ---\n`);
		}
		this.connected = false;
		this.attached = false;
		this.log('\n--- Disconnected ---\n');
	};
}

declare global {
	// eslint-disable-next-line no-var
	var __webSerialConnection: WebSerialConnection | undefined;
}

export function getWebSerialConnection() {
	globalThis.__webSerialConnection ??= new WebSerialConnection();
	return globalThis.__webSerialConnection;
}
