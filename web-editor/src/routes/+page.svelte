<script lang="ts">
	import { getWebSerialConnection, type WebSerialConnection } from '$lib/connection.svelte';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Resizable from '$lib/components/ui/resizable';
	import { Input } from '$lib/components/ui/input/index';

	let outputelement: HTMLPreElement | null = $state(null);

	$effect(() => {
		const output = connection?.output;
		if (outputelement && connection && connection.connected) {
			void output;
			outputelement.scrollTop = outputelement.scrollHeight;
		}
	});

	let terminalInput = $state('');
	function sendTerminalInput() {
		if (connection && connection.connected) {
			connection.sendText(terminalInput + '\r');
			terminalInput = '';
		} else {
			alert('connect before sending.');
		}
	}

	async function saveToDevice() {
		if (!connection || !connection.connected) {
			alert('connect before sending.');
			return;
		}
		await connection.saveToDevice(editorContent);
	}

	function getSavedScript() {
		const savedScript = localStorage.getItem('editorContent');
		if (savedScript) {
			editorContent = savedScript;
		}
	}

	function handleSubmit(event: Event) {
		event.preventDefault();
		sendTerminalInput();
	}

	function save() {
		localStorage.setItem('editorContent', editorContent);
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
		getSavedScript();
	});

	async function uploadScript() {
		if (!connection || !connection.connected) {
			alert('Not connected to the device.');
			return;
		}
		await connection.runScript(editorContent);
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
				<Button onclick={saveToDevice}>Save to device</Button>
				<Button variant="secondary" onclick={save}>Save script</Button>
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
				<div class="flex flex-col h-full w-full min-h-0">
					{#if connection.connected}
						<pre
							class="flex-1 min-h-0 overflow-y-auto whitespace-pre-wrap"
							bind:this={outputelement}>{connection.output}</pre>
					{:else}
						<p class="flex-1 min-h-0 overflow-y-auto">
							Connect to the device to see output and send terminal commands.
						</p>
					{/if}
					<form onsubmit={handleSubmit} class="w-full flex-row flex justify-between">
						<Input
							class="w-full"
							bind:value={terminalInput}
							disabled={!connection.connected}
							placeholder="terminal input"
						/>
						<Button type="submit" disabled={!connection.connected}>Send</Button>
					</form>
				</div>
			</Resizable.Pane>
		</Resizable.PaneGroup>
	</div>
{/if}
