<script lang="ts">
	let { editorContent = $bindable() } = $props();
	//You can bind to editor content, but it is read only (from the pov of the parent component) once it's initialised.
	import { basicSetup } from 'codemirror';
	import { onMount } from 'svelte';
	let editorContainer: HTMLDivElement | null = $state(null);
	import { EditorState } from '@codemirror/state';
	import { python } from '@codemirror/lang-python';

	import {
		EditorView,
		keymap,
		highlightSpecialChars,
		drawSelection,
		highlightActiveLine,
		dropCursor,
		rectangularSelection,
		crosshairCursor,
		lineNumbers,
		highlightActiveLineGutter
	} from '@codemirror/view';
	import {
		defaultHighlightStyle,
		syntaxHighlighting,
		indentOnInput,
		bracketMatching,
		foldGutter,
		foldKeymap
	} from '@codemirror/language';
	import { defaultKeymap, history, historyKeymap, indentWithTab } from '@codemirror/commands';
	import { searchKeymap, highlightSelectionMatches } from '@codemirror/search';
	import {
		autocompletion,
		completionKeymap,
		closeBrackets,
		closeBracketsKeymap
	} from '@codemirror/autocomplete';
	import { lintKeymap } from '@codemirror/lint';

	onMount(() => {
		const view = new EditorView({
			doc: editorContent,
			parent: editorContainer,
			extensions: [
				EditorView.updateListener.of((update: any) => {
					// Ensure the update actually changed the document content
					if (update.docChanged) {
						editorContent = view.state.doc.toString();
						console.log('detected change.')
					}
				}),
				// A line number gutter
				lineNumbers(),
				python(),
				// A gutter with code folding markers
				foldGutter(),
				// Replace non-printable characters with placeholders
				highlightSpecialChars(),
				// The undo history
				history(),
				// Replace native cursor/selection with our own
				drawSelection(),
				// Show a drop cursor when dragging over the editor
				dropCursor(),
				// Allow multiple cursors/selections
				EditorState.allowMultipleSelections.of(true),
				// Re-indent lines when typing specific input
				indentOnInput(),
				// Highlight syntax with a default style
				syntaxHighlighting(defaultHighlightStyle),
				// Highlight matching brackets near cursor
				bracketMatching(),
				// Automatically close brackets
				closeBrackets(),
				// Load the autocompletion system
				autocompletion(),
				// Allow alt-drag to select rectangular regions
				rectangularSelection(),
				// Change the cursor to a crosshair when holding alt
				crosshairCursor(),
				// Style the current line specially
				highlightActiveLine(),
				// Style the gutter for current line specially
				highlightActiveLineGutter(),
				// Highlight text that matches the selected text
				highlightSelectionMatches(),
				keymap.of([
					// Closed-brackets aware backspace
					...closeBracketsKeymap,
					indentWithTab,
					// A large set of basic bindings
					...defaultKeymap,
					// Search-related keys
					...searchKeymap,
					// Redo/undo keys
					...historyKeymap,
					// Code folding bindings
					...foldKeymap,
					// Autocompletion keys
					...completionKeymap,
					// Keys related to the linter system
					...lintKeymap
				])
			]
		});
	});
</script>

<!-- nested divs are a plague upon this  -->
<div class="w-full h-full">
	<div class="editor h-full" bind:this={editorContainer}></div>
</div>

<style>
	.editor :global(.cm-editor) {
		height: 100%;
	}
</style>
