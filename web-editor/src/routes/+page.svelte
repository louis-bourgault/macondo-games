<script lang="ts">
	let port: any;
	let reader: any;
	let writer: any;
	let keepreading = false;
	let readableStreamClosed;
	let connected = $state(false);
    let output = $state('');
	let command = $state('');
	async function connect() {
		if (!('serial' in navigator)) {
			alert('Web Serial API not supported. Use Chrome or Edge over HTTPS (or localhost).');
			return;
		}
		port = await navigator.serial.requestPort();
		const baudRate = 115200;
		await port.open({ baudRate });
		connected = true;

		const textDecoder = new TextDecoderStream();
		readableStreamClosed = port.readable.pipeTo(textDecoder.writable);
		reader = textDecoder.readable.getReader();

		const textEncoder = new TextEncoderStream();
		textEncoder.readable.pipeTo(port.writable);
		writer = textEncoder.writable.getWriter();

		keepreading = true;
		readLoop();
		console.log('connected to serial port');
	}

	function controlD() {
		sendText('\x04');
	}

	function controlC() {
		sendText('\x03');
	}

	function log(msg: string) {
        output += msg;
		console.log(msg);
	}

	async function sendText(text: string) {
		if (!writer) return;
		await writer.write(text);
	}

	async function readLoop() {
		try {
			while (keepreading) {
				const { value, done } = await reader.read();
				if (done) {
					reader.releaseLock();
					break;
				}
				if (value) log(value);
			}
		} catch (err: any) {
			log(`\n--- Read error: ${err.message} ---\n`);
		}
	}

	async function disconnect() {
		keepreading = false;
		try {
			if (reader) {
				await reader.cancel();
				reader = null;
			}
			if (writer) {
				await writer.close();
				writer = null;
			}
			if (port) {
				await port.close();
				port = null;
			}
		} catch (err: any) {
			log(`\n--- Disconnect error: ${err.message} ---\n`);
		}
		connected = false;
		log('\n--- Disconnected ---\n');
	}
</script>

<div class="h-10 w-full flex flex-row">
	<button class="bg-amber-950 text-white p-2" onclick={controlC}>Send control c</button>
    <button class="bg-amber-950 text-white p-2" onclick={controlD}>Send control d</button>
	<button class="bg-green-950 text-white p-2" onclick={connect}>Connect</button>
</div>

<div class='output h-40 w-full bg-amber-50 border-2 border-amber-950 p-2 overflow-y-auto'>
    <pre id='output'>{output}</pre>
</div>

<div class='command h-20 w-full flex flex-col'>
	<input bind:value={command} class='w-full h-10 p-2 border-2 border-amber-950' placeholder='send to terminal' />
	<button class='bg-amber-950 text-white p-2' onclick={() => sendText(command)}>Send</button>
</div>