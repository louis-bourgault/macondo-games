<script lang="ts">
	import { getWebSerialConnection, type WebSerialConnection } from '$lib/connection.svelte';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Resizable from '$lib/components/ui/resizable';
	import { Input } from '$lib/components/ui/input/index';

	let terminalInput = $state('');
	function sendTerminalInput() {
		if (connection && connection.connected) {
			connection.sendText(terminalInput + '\r');
			terminalInput = '';
		} else {
			alert('connect before sending.');
		}
	}

	function handleSubmit(event: Event) {
		event.preventDefault();
		sendTerminalInput();
	}

	let editorContent = $state(`from modules.display import Display
from modules.input import Input

display = Display()
display.fill(0x0000) 
display.update()
    `);
	let connection = $state<WebSerialConnection | null>(null);
	onMount(() => {
		connection = getWebSerialConnection();
		void connection.init();
	});

	function uploadScript() {
		const script = editorContent;
		if (connection && connection.connected) {
			connection.controlD(); //soft reboot, neccesary to clear display buffer memory allocation
			connection.controlC(); //stop running script
			connection.sendText(script + '\r');
		} else {
			alert('Not connected to the device.');
		}
	}
</script>

{#if connection}
	<div class="h-full w-full flex flex-col overscroll-none">
		<div class="h-5 w-full flex flex-row">
			<Button onclick={connection.connect}>Connect</Button>
			{#if connection.connected}
				<Button variant="secondary" onclick={connection.controlC}>Send control c</Button>
				<Button variant="secondary" onclick={connection.controlD}>Send control d</Button>
				<Button variant="secondary" onclick={connection.disconnect}>Disconnect</Button>
				<Button onclick={uploadScript}>Run Script</Button>
			{/if}
		</div>
		<br class="h-1" />
		<Resizable.PaneGroup
			direction="vertical"
			class="min-h-50 max-w-md rounded-lg border min-w-screen"
		>
			<Resizable.Pane defaultSize={70} minSize={30}>
				<textarea bind:value={editorContent} class="w-full h-full p-2 border-2 border-amber-950"
				></textarea>
			</Resizable.Pane>
			<Resizable.Handle />
			<Resizable.Pane defaultSize={30} minSize={10}>
				<div class="flex flex-col h-full w-full">
					<pre class="overflow-scroll">{connection.output}</pre>
						<form onsubmit={handleSubmit} class='w-full flex-row flex justify-between'>
							<Input class="w-full" bind:value={terminalInput} placeholder="terminal input" />
							<Button type="submit">Send</Button>
						</form>
				</div>
			</Resizable.Pane>
		</Resizable.PaneGroup>
	</div>
{/if}
