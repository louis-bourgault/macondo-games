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
		let hash = 0x811C9DC5; 
		
		for (let i = 0; i < str.length; i++) {
			hash ^= str.charCodeAt(i);
			hash = Math.imul(hash, 0x01000193);
		}
		
		return (hash >>> 0).toString(16).padStart(8, '0');
	}

	public async syncImages(imgs: ImageFile[]) {
		console.log('received images to sync', imgs);

		let sentImagesData = '';

		//request the file with all the checksums
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.controlE(); //paste mode
		await this.delay(50);
		await this.sendText(`try:\n    with open('/img/images.txt', 'r') as f:\n        _data = f.read()\nexcept OSError:\n    _data = ''\nprint(_data)\n`);
		await this.controlD();
		await this.waitForPrompt();

		const existingImagesData = this.output.split('>>> ')[1].trim();
		console.log('existing images data', existingImagesData);

		const existingImagesMap = new Map<string, { width: number; height: number; checksum: string }>();
		for (const line of existingImagesData.split('\n')) {
			const [name, widthStr, heightStr, checksum] = line.split(',');
			if (name && widthStr && heightStr && checksum) {
				existingImagesMap.set(name, {
					width: parseInt(widthStr, 10),
					height: parseInt(heightStr, 10),
					checksum,
				});
			}
		}



		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.waitForPrompt();

		const chunkSize = 8192; //8kb seems to be a decent chunk size to balance speed and stability.
		//if its too big,we OOM, if its too small, it takes forever to send.

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
					sentImagesData += `\n${img.name},${img.width},${img.height},${checksum}`;
					continue;
				}
			}
			sentImagesData += `\n${img.name},${img.width},${img.height},${checksum}`;
			
			await this.controlE();
			await this.delay(50);
			await this.sendText(
				`import os, binascii\ntry:\n    os.mkdir('/img')\nexcept OSError:\n    pass\n_f = open('/img/${img.name}', 'wb')\n`
			);
			await this.controlD();
			await this.waitForPrompt();
			
			//seperate paste blocks so we don't OOM
			for (let i = 0; i < img.content.length; i += chunkSize) {
				const chunk = img.content.slice(i, i + chunkSize);
				await this.controlE();
				await this.delay(50);
				await this.sendText(`_f.write(binascii.a2b_base64('${chunk}'))\n`);
				await this.controlD();
				await this.waitForPrompt();
			}

			await this.controlE();
			await this.delay(50);
			await this.sendText(`_f.close()\n`);
			await this.controlD();
			await this.waitForPrompt();
		}

		for (const [name, { width, height, checksum }] of existingImagesMap.entries()) {
			if (!imgs.find((img) => img.name === name)) {
				//delete the image from the device if it is not in the new list
				await this.controlE();
				await this.delay(50);
				await this.sendText(`import os\ntry:\n    os.remove('/img/${name}')\nexcept OSError:\n    pass\n`);
				await this.controlD();
				await this.waitForPrompt();
			}
		}

		await this.controlE();
		await this.delay(50);
		await this.sendText(`with open('/img/images.txt', 'w') as f:\n    f.write("""${sentImagesData}""")\n`);
		await this.controlD();
		await this.waitForPrompt();
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
	}

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
	};

	public writeFileToDevice = async (filename: string, content: string) => {
		await this.controlC();
		await this.delay(100);
		await this.controlC();
		await this.delay(100);
		await this.controlE();
		await this.delay(50);
		await this.sendText(`with open('${filename}', 'w') as f:\n    f.write("""${content}""")\n`);
		await this.controlD();
		await this.delay(500);
		await this.controlD();
	};

	public runProject = async (projectData: ProjectData) => {
		const mainFile =
			projectData.codeFiles.find((f) => f.name === 'main.py') ?? projectData.codeFiles[0];
		// Write every non-main file to the device so `import` statements in the
		// main file resolve at runtime. The main file itself is executed inline
		// via paste mode (see runScript), so it does not need to be persisted.
		for (const file of projectData.codeFiles) {
			if (file.name === mainFile.name) continue;
			await this.writeFileToDevice(file.name, file.content);
		}
		await this.runScript(mainFile.content);
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
