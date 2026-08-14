import type { ProjectData, ImageFile } from './types';

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

	private async waitForPrompt(timeoutMs = 5000): Promise<void> {
		const startLen = this.output.length;
		const deadline = Date.now() + timeoutMs;
		while (Date.now() < deadline) {
			if (this.output.slice(startLen).includes('>>> ')) return;
			await this.delay(10);
		}
		throw new Error('timeout waiting for >>> prompt');
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

	public async syncImages(imgs: ImageFile[]) {
		console.log('received images to sync', imgs);

		// Stop whatever is running so the REPL is responsive, then read the
		// image manifest that the firmware keeps on its own filesystem.
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		await this.controlE(); //paste mode
		await this.delay(50);
		// Note where output ends before querying the manifest so we only parse
		// the response to THIS query. connection.output accumulates the whole
		// session (including the controlC prompts above), so splitting on the
		// first '>>> ' would grab the stale bytes between the first two prompts.
		const manifestStart = this.output.length;
		await this.sendText(`import ferret\nprint(ferret.image_manifest())\n`);
		await this.controlD();
		await this.waitForPrompt();

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

		const chunkSize = 2048; //must fit the MP GC heap (16 KiB): a base64 string literal this big plus
		//its parse tree/bytecode OOMs at ~4 KiB, so 2 KiB is a safe margin. Too big and the paste
		//fails to compile; too small and it takes forever to send.

		for (const img of imgs) {
			if (!img.name || !img.content) {
				console.error("can't send this one", img);
				continue;
			}
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

			// Upload in chunks: write_image opens the upload, append_image adds
			// base64 characters (chunks may split anywhere; the device decodes
			// the accumulated payload once), write_image_end validates the size
			// and persists the image plus its manifest entry.
			for (let i = 0; i < img.content.length; i += chunkSize) {
				const chunk = img.content.slice(i, i + chunkSize);
				const call =
					i === 0
						? `ferret.write_image("${img.name}", ${img.width}, ${img.height}, "${chunk}")`
						: `ferret.append_image("${img.name}", "${chunk}")`;
				//seperate paste blocks so we don't OOM
				await this.controlE();
				await this.delay(50);
				await this.sendText(`import ferret\n${call}\n`);
				await this.controlD();
				await this.waitForPrompt();
			}

			await this.controlE();
			await this.delay(50);
			await this.sendText(`import ferret\nferret.write_image_end("${img.name}")\n`);
			await this.controlD();
			await this.waitForPrompt();
		}

		for (const name of existingImagesMap.keys()) {
			if (!imgs.find((img) => img.name === name)) {
				//delete the image from the device if it is not in the new list
				await this.controlE();
				await this.delay(50);
				await this.sendText(
					`import ferret\ntry:\n    ferret.delete_image("${name}")\nexcept OSError:\n    pass\n`
				);
				await this.controlD();
				await this.waitForPrompt();
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
		await this.controlE();
		await this.delay(50);
		await this.sendText(
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
		await this.controlD();
	};

	public saveProjectToDevice = async (projectData: ProjectData) => {
		for (const file of projectData.codeFiles) {
			await this.writeFileToDevice(file.name, file.content);
		}
		// Boot main.py right away: a soft reset re-runs the boot script, so the
		// game starts immediately instead of needing a power cycle.
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		await this.controlD();
	};

	public writeFileToDevice = async (filename: string, content: string) => {
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);

		// Chunked so we don't OOM the MicroPython heap. write_file opens the
		// file, append_file adds the rest. Backslashes and triple quotes are
		// escaped so the paste literal round-trips the content byte-for-byte.
		// 2 KiB keeps the string literal well under the 16 KiB GC heap (4 KiB
		// chunks already fail to compile).
		const chunkSize = 2048;
		for (let i = 0; i < content.length; i += chunkSize) {
			const chunk = content
				.slice(i, i + chunkSize)
				.replace(/\\/g, '\\\\')
				.replace(/"""/g, '\\"""');
			const call =
				i === 0
					? `ferret.write_file("${filename}", """${chunk}""")`
					: `ferret.append_file("${filename}", """${chunk}""")`;
			await this.controlE();
			await this.delay(50);
			await this.sendText(`import ferret\n${call}\n`);
			await this.controlD();
			await this.waitForPrompt();
		}
	};

	public runProject = async (projectData: ProjectData) => {
		// Persist every file (including main.py) so `import` statements resolve
		// and the boot script has fresh content, then soft-reset: main.py runs
		// automatically from flash with a fresh namespace.
		for (const file of projectData.codeFiles) {
			await this.writeFileToDevice(file.name, file.content);
		}
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
