import { type ProjectData } from './types';

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
