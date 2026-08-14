<script lang="ts">
	let { editorContent = $bindable() } = $props();
	//You can bind to editor content, but it is read only (from the pov of the parent component) once it's initialised.
	import { onMount } from 'svelte';
	let editorContainer: HTMLDivElement | null = $state(null);
	import { EditorState } from '@codemirror/state';
	import { python, pythonLanguage } from '@codemirror/lang-python';
	import { tags } from '@lezer/highlight';

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
		HighlightStyle,
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
		closeBracketsKeymap,
		type Completion,
		type CompletionContext,
		type CompletionResult
	} from '@codemirror/autocomplete';
	import { lintKeymap } from '@codemirror/lint';

	const libraryCompletions: Completion[] = [
		{
			label: 'ferret.fill',
			type: 'function',
			detail: '(color: int)',
			info: 'Fill the whole display with a single RGB565 color, e.g. ferret.rgb_to_565(255, 0, 0) for red.'
		},
		{
			label: 'ferret.fill_rect',
			type: 'function',
			detail: '(x, y, width, height, color)',
			info: 'Draw a filled rectangle in the display buffer.'
		},
		{
			label: 'ferret.pixel',
			type: 'function',
			detail: '(x, y, color)',
			info: 'Set a single pixel in the display buffer.'
		},
		{
			label: 'ferret.present',
			type: 'function',
			detail: '()',
			info: 'Send whatever is in the display buffer to the actual display. Call this after drawing.'
		},
		{
			label: 'ferret.draw_text',
			type: 'function',
			detail: '(text, x, y, color)',
			info: 'Draw text at the given position. Use ferret.measure_text to center it.'
		},
		{
			label: 'ferret.measure_text',
			type: 'function',
			detail: '(text) -> (width, height)',
			info: 'Measure the space text will take up, so you can position or center it.'
		},
		{
			label: 'ferret.rgb_to_565',
			type: 'function',
			detail: '(r, g, b) -> int',
			info: 'Convert an 8-bit-per-channel color to the RGB565 format the display uses.'
		},
		{
			label: 'ferret.draw_image',
			type: 'function',
			detail: '(name, x, y)',
			info: 'Draw an image from the device at (x, y). Its size is read from the image manifest, so sync your images first.'
		},
		{
			label: 'ferret.random_int',
			type: 'function',
			detail: '(min, max) -> int',
			info: 'A random integer in the range [min, max], e.g. ferret.random_int(1, 6) rolls a die.'
		},
		{
			label: 'ferret.input_update',
			type: 'function',
			detail: '()',
			info: 'Update the input state. Call this once per frame before checking buttons.'
		},
		{
			label: 'ferret.input_is_pressed',
			type: 'function',
			detail: '(button: str) -> bool',
			info: 'Check if a button is currently held down. Call ferret.input_update() first. Buttons: "A", "B", "UP", "DOWN", "LEFT", "RIGHT", "START", "EXIT".'
		},
		{
			label: 'ferret.input_was_just_pressed',
			type: 'function',
			detail: '(button: str) -> bool',
			info: 'Check if a button was pressed since the last call to ferret.input_update(). Call ferret.input_update() first. Buttons: "A", "B", "UP", "DOWN", "LEFT", "RIGHT", "START", "EXIT".'
		},
		{
			label: 'ferret.input_was_just_released',
			type: 'function',
			detail: '(button: str) -> bool',
			info: 'Check if a button was released since the last call to ferret.input_update(). Call ferret.input_update() first. Buttons: "A", "B", "UP", "DOWN", "LEFT", "RIGHT", "START", "EXIT".'
		}
		/*
		// Internal ferret API used by the editor (save/sync) — not for student
		// code, so kept out of the completions. Uncomment to bring them back.
		{
			label: 'ferret.write_file',
			type: 'function',
			detail: '(path, content)',
			info: 'Write a file to the device flash (used by the editor when saving, but handy for things like saving a high score).'
		},
		{
			label: 'ferret.append_file',
			type: 'function',
			detail: '(path, content)',
			info: 'Append content to a file on the device flash. Create it first with ferret.write_file.'
		},
		{
			label: 'ferret.image_manifest',
			type: 'function',
			detail: '() -> str',
			info: 'List the images stored on the device, one "name,width,height,checksum" per line. Used by the editor when syncing images.'
		},
		{
			label: 'ferret.delete_image',
			type: 'function',
			detail: '(name)',
			info: 'Remove an image from the device flash. Raises OSError if it does not exist.'
		},
		{
			label: 'ferret.write_image',
			type: 'function',
			detail: '(name, width, height, base64_chunk)',
			info: 'Start uploading an image as base64 (used by the editor). Finish with ferret.write_image_end.'
		},
		{
			label: 'ferret.append_image',
			type: 'function',
			detail: '(name, base64_chunk)',
			info: 'Add more base64 characters to an in-progress image upload (used by the editor).'
		},
		{
			label: 'ferret.write_image_end',
			type: 'function',
			detail: '(name)',
			info: 'Validate and save a finished image upload (used by the editor).'
		}
		*/
	];

	function libraryCompletionSource(context: CompletionContext): CompletionResult | null {
		const word = context.matchBefore(/[\w.]+/);
		if (!word && !context.explicit) return null;
		return {
			from: word ? word.from : context.pos,
			options: libraryCompletions,
			validFor: /^[\w.]*$/
		};
	}

	let view: EditorView;

	const libraryCompletionExtension = pythonLanguage.data.of({
		autocomplete: libraryCompletionSource
	});

	// A CodeMirror theme built on shadcn's CSS variables. Because the
	// variables are redefined under `.dark` (toggled by mode-watcher on
	// the <html> element), the editor restyles itself automatically when
	// dark mode is switched on or off — no extension reconfiguration
	// needed. The selectors are scoped by EditorView.theme() so they
	// override the built-in &light/&dark base-theme rules.
	const editorTheme = EditorView.theme({
		'&': {
			backgroundColor: 'var(--background)',
			color: 'var(--foreground)',
			height: '100%',
			fontSize: '14px'
		},
		'.cm-scroller': {
			fontFamily:
				'var(--font-mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace)',
			lineHeight: '1.6'
		},
		'.cm-content': {
			caretColor: 'var(--foreground)',
			maxWidth: 'none'
		},
		'.cm-cursor, .cm-dropCursor': {
			borderLeftColor: 'var(--foreground)'
		},
		'&.cm-focused .cm-cursor': {
			borderLeftColor: 'var(--primary)'
		},
		'.cm-selectionBackground': {
			backgroundColor: 'color-mix(in oklch, var(--accent) 70%, transparent)'
		},
		'&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground': {
			backgroundColor: 'color-mix(in oklch, var(--accent) 80%, transparent)'
		},
		'::selection': {
			backgroundColor: 'color-mix(in oklch, var(--accent) 70%, transparent)'
		},
		'.cm-activeLine': {
			backgroundColor: 'color-mix(in oklch, var(--muted) 55%, transparent)'
		},
		'.cm-activeLineGutter': {
			backgroundColor: 'color-mix(in oklch, var(--muted) 55%, transparent)',
			color: 'var(--foreground)'
		},
		'.cm-gutters': {
			backgroundColor: 'transparent',
			color: 'var(--muted-foreground)',
			border: 'none'
		},
		'.cm-gutterElement': {
			color: 'var(--muted-foreground)'
		},
		'.cm-foldPlaceholder': {
			backgroundColor: 'var(--secondary)',
			color: 'var(--muted-foreground)',
			border: '1px solid var(--border)'
		},
		'.cm-specialChar': {
			color: 'var(--destructive)'
		},
		'.cm-panels': {
			backgroundColor: 'var(--card)',
			color: 'var(--card-foreground)'
		},
		'.cm-panels-top': {
			borderBottom: '1px solid var(--border)'
		},
		'.cm-panels-bottom': {
			borderTop: '1px solid var(--border)'
		},
		'.cm-textfield, .cm-dialog input': {
			backgroundColor: 'var(--input)',
			color: 'var(--foreground)',
			border: '1px solid var(--border)'
		},
		'.cm-button': {
			backgroundColor: 'var(--secondary)',
			color: 'var(--secondary-foreground)',
			border: '1px solid var(--border)'
		},
		'.cm-tooltip': {
			backgroundColor: 'var(--popover)',
			color: 'var(--popover-foreground)',
			border: '1px solid var(--border)'
		},
		'.cm-tooltip .cm-tooltip-arrow': {
			color: 'var(--border)'
		},
		'.cm-tooltip-autocomplete > ul > li[aria-selected]': {
			backgroundColor: 'var(--accent)',
			color: 'var(--accent-foreground)'
		},
		'.cm-tooltip-autocomplete .cm-completionDetail': {
			color: 'var(--muted-foreground)'
		},
		'&.cm-editor.cm-focused': {
			outline: 'none'
		}
	});

	// Syntax highlighting that also uses CSS variables, so token colors
	// swap between the light and dark palettes defined in layout.css.
	const editorHighlightStyle = HighlightStyle.define([
		{ tag: tags.comment, color: 'var(--syntax-comment)', fontStyle: 'italic' },
		{ tag: tags.lineComment, color: 'var(--syntax-comment)', fontStyle: 'italic' },
		{ tag: tags.blockComment, color: 'var(--syntax-comment)', fontStyle: 'italic' },
		{ tag: tags.name, color: 'var(--syntax-variable)' },
		{ tag: tags.variableName, color: 'var(--syntax-variable)' },
		{ tag: tags.definition(tags.variableName), color: 'var(--syntax-variable)' },
		{ tag: tags.local(tags.variableName), color: 'var(--syntax-variable)' },
		{ tag: tags.special(tags.variableName), color: 'var(--syntax-variable)' },
		{ tag: tags.constant(tags.variableName), color: 'var(--syntax-constant)' },
		{ tag: tags.function(tags.variableName), color: 'var(--syntax-function)' },
		{ tag: tags.propertyName, color: 'var(--syntax-property)' },
		{ tag: tags.function(tags.propertyName), color: 'var(--syntax-function)' },
		{ tag: tags.typeName, color: 'var(--syntax-type)' },
		{ tag: tags.className, color: 'var(--syntax-type)' },
		{ tag: tags.namespace, color: 'var(--syntax-namespace)' },
		{ tag: tags.macroName, color: 'var(--syntax-macro)' },
		{ tag: tags.labelName, color: 'var(--syntax-variable)' },
		{ tag: tags.standard(tags.name), color: 'var(--syntax-builtin)' },
		{ tag: tags.keyword, color: 'var(--syntax-keyword)' },
		{ tag: tags.controlKeyword, color: 'var(--syntax-controlKeyword)' },
		{ tag: tags.operatorKeyword, color: 'var(--syntax-operatorKeyword)' },
		{ tag: tags.definitionKeyword, color: 'var(--syntax-definitionKeyword)' },
		{ tag: tags.moduleKeyword, color: 'var(--syntax-moduleKeyword)' },
		{ tag: tags.modifier, color: 'var(--syntax-modifier)' },
		{ tag: tags.atom, color: 'var(--syntax-atom)' },
		{ tag: tags.bool, color: 'var(--syntax-bool)' },
		{ tag: tags.null, color: 'var(--syntax-null)' },
		{ tag: tags.number, color: 'var(--syntax-number)' },
		{ tag: tags.integer, color: 'var(--syntax-number)' },
		{ tag: tags.float, color: 'var(--syntax-number)' },
		{ tag: tags.string, color: 'var(--syntax-string)' },
		{ tag: tags.character, color: 'var(--syntax-string)' },
		{ tag: tags.regexp, color: 'var(--syntax-regexp)' },
		{ tag: tags.escape, color: 'var(--syntax-escape)' },
		{ tag: tags.operator, color: 'var(--syntax-operator)' },
		{ tag: tags.derefOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.arithmeticOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.logicOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.bitwiseOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.compareOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.updateOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.definitionOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.typeOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.controlOperator, color: 'var(--syntax-operator)' },
		{ tag: tags.punctuation, color: 'var(--syntax-punctuation)' },
		{ tag: tags.separator, color: 'var(--syntax-punctuation)' },
		{ tag: tags.bracket, color: 'var(--syntax-punctuation)' },
		{ tag: tags.meta, color: 'var(--syntax-meta)' },
		{ tag: tags.processingInstruction, color: 'var(--syntax-meta)' },
		{ tag: tags.invalid, color: 'var(--syntax-invalid)' }
	]);

	const extentions = [
		editorTheme,
		EditorView.updateListener.of((update: any) => {
			// Ensure the update actually changed the document content
			if (update.docChanged) {
				editorContent = view.state.doc.toString();
				console.log('detected change.');
			}
		}),
		// A line number gutter
		lineNumbers(),
		python(),
		libraryCompletionExtension,
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
		// Highlight syntax using our variable-driven theme
		syntaxHighlighting(editorHighlightStyle),
		// Fall back to the default style for any tokens we didn't cover
		syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
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
	];

	export function loadNewEditorContent(newContent: string) {
		const newState = EditorState.create({
			doc: newContent,
			extensions: extentions
		});
		view.setState(newState);
	}

	onMount(() => {
		view = new EditorView({
			doc: editorContent,
			parent: editorContainer!,
			extensions: extentions
		});
	});
</script>

<!-- nested divs are a plague upon this  -->
<div class="h-full w-full">
	<div class="editor h-full" bind:this={editorContainer}></div>
</div>

<style>
	.editor :global(.cm-editor) {
		height: 100%;
	}
</style>
