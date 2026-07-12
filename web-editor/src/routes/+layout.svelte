<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.png';
	import { ModeWatcher } from "mode-watcher";
	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} />
<title>Editor | Ferretboard</title></svelte:head>
<ModeWatcher/>
<div class='h-full w-full'>
	{@render children()}
</div>
