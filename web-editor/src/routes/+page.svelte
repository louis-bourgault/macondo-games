<script lang="ts">
	import { WebSerialConnection } from '$lib/connection.svelte';
	import { onMount } from 'svelte';

	let command = $state('');
	let connection = $state<WebSerialConnection | null>(null);

	onMount(() => {
		connection = new WebSerialConnection();
	});
</script>

{#if !connection}
	<div class="h-full w-full flex flex-col items-center justify-center">
		<p class="text-amber-950">Loading</p>
		<p>Loading</p>
	</div>
{:else}
	<div class="h-10 w-full flex flex-row">
		<button class="bg-amber-950 text-white p-2" onclick={connection.controlC}>Send control c</button
		>
		<button class="bg-amber-950 text-white p-2" onclick={connection.controlD}>Send control d</button
		>
		<button class="bg-green-950 text-white p-2" onclick={connection.connect}>Connect</button>
	</div>

	<div class="output h-40 w-full bg-amber-50 border-2 border-amber-950 p-2 overflow-y-auto">
		<pre id="output">{connection.output}</pre>
	</div>

	<div class="command h-20 w-full flex flex-col">
		<input
			bind:value={command}
			class="w-full h-10 p-2 border-2 border-amber-950"
			placeholder="send to terminal"
		/>
		<button class="bg-amber-950 text-white p-2" onclick={() => connection.sendText(command + '\r')}
			>Send</button
		>
	</div>
{/if}
