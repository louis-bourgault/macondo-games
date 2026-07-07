export class WebSerialConnection {
    public port: any;
	public reader: any;
	public writer: any;
	public keepreading = false;
	public readableStreamClosed: any;
	public connected = $state(false);
    public output = $state('');
	private attached = false;

	private async attachPort(port: any) {
		if (this.attached && this.port === port) {
			this.connected = true;
			return;
		}

		this.port = port;
		if (this.port.readable?.locked || this.port.writable?.locked) {
			this.connected = true;
			this.attached = true;
			return;
		}
		const textDecoder = new TextDecoderStream();
		this.readableStreamClosed = this.port.readable.pipeTo(textDecoder.writable);
		this.reader = textDecoder.readable.getReader();

		const textEncoder = new TextEncoderStream();
		textEncoder.readable.pipeTo(this.port.writable);
		this.writer = textEncoder.writable.getWriter();

		this.keepreading = true;
		void this.readLoop();
		this.connected = true;
		this.attached = true;
	}

	public init = async () => {
		const serial = (navigator as Navigator & { serial?: any }).serial;
		if (!serial) return;

		const ports = await serial.getPorts();
		const port = ports[0];
		if (!port) return;

		if (port.readable && port.writable) {
			await this.attachPort(port);
		}
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
		this.port = await serial.requestPort();
		const baudRate = 115200;
		await this.port.open({ baudRate });
		await this.attachPort(this.port);
		console.log('connected to serial port');
	};

	public controlD = () => {
        this.output = '';
		this.sendText('\x04');
	};

	public controlC = () => {
		this.sendText('\x03');
	};

	public sendText = async (text: string) => {
		if (!this.writer) return;
		await this.writer.write(text);
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