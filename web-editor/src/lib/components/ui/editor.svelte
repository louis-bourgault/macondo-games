<script lang="ts">
	import { basicSetup } from 'codemirror';
	import { EditorView } from '@codemirror/view';
    import {onMount} from 'svelte';
    let editorContainer: HTMLDivElement | null = $state(null);

	onMount(() => {
        const view = new EditorView({
            doc: 'Start document',
            parent: editorContainer,
            extensions: [basicSetup]
        });
    })
</script>

<div class="editor-container">
    <div class="editor" bind:this={editorContainer}></div>
</div>
