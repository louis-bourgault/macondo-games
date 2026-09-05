<script lang="ts">
	import {
		getWebSerialConnection,
		isValidCodeFileName,
		type WebSerialConnection
	} from '$lib/connection.svelte';
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Resizable from '$lib/components/ui/resizable';
	import { Input } from '$lib/components/ui/input/index';
	import Editor from '$lib/components/Editor.svelte';
	import ResizableHandle from '$lib/components/ui/resizable/resizable-handle.svelte';
	import * as Accordion from '$lib/components/ui/accordion/index.js';
	import SidebarDocs from '$lib/components/SidebarDocs.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	let outputelement: HTMLPreElement | null = $state(null);
	let editor: Editor | null = $state(null);
	import { type ProjectData } from '$lib/types';
	import ImageEditor from '$lib/components/ImageEditor.svelte';
	import {
		FilePlus,
		FileCode,
		ImagePlus,
		Image as ImageIcon,
		PencilLine,
		Trash,
		TriangleAlert
	} from '@lucide/svelte';

	let projectData: ProjectData = $state({
		name: 'My Project',
		codeFiles: [
			{
				name: 'main.py',
				content: `import ferret\n\n# Your code here!`
			}
		],
		images: []
	});

	async function syncImages() {
		console.log('syncing images...');
		if (!connection || !connection.connected) {
			alert('connect before sending.');
			return;
		}
		try {
			await connection.syncImages(projectData.images);
		} catch (err) {
			console.error('image sync failed', err);
			alert(`Image sync failed: ${err instanceof Error ? err.message : String(err)}`);
		}
	}

	let selectedImageIndex: number | null = $state(null);

	let currentFileIndex = $state(0);

	function changeActiveFile(index: number) {
		currentFileIndex = index;
		//the bind: directive is single direction, so we need to call this function to update the editor content when the current file changes.
		//i hope that the svelte reactivity will change the editor index before this function is called, so that we dont save the wrong thing in the interim

		//UPDATE: it looks like this works fine; ie svelte reactivity updates the editor before this function is called.
		editor?.loadNewEditorContent(projectData.codeFiles[currentFileIndex].content);
	}

	// --- file / image management ---
	const MAX_IMAGE_DIM = 240;
	const MAIN_FILE = 'main.py';

	function imageNameExists(name: string, excludeIndex: number | null = null) {
		return projectData.images.some((img, i) => i !== excludeIndex && img.name === name);
	}
	function imageNameIsValid(name: string) {
		return name !== '' && name !== '.' && name !== '..' && !/[\\/:,\r\n]/.test(name);
	}
	function fileNameExists(name: string, excludeIndex: number | null = null) {
		return projectData.codeFiles.some((f, i) => i !== excludeIndex && f.name === name);
	}
	function fileNameError(name: string) {
		return isValidCodeFileName(name)
			? ''
			: 'Use a Python identifier ending in .py, for example player.py.';
	}

	// dialogs
	let newImageDialogOpen = $state(false);
	let newFileDialogOpen = $state(false);
	let renameDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);

	// new item forms
	let newImageData = $state({ name: '', width: 0, height: 0 });
	let newImageError = $state('');
	let newFileName = $state('');
	let newFileError = $state('');

	// rename (shared for files + images)
	let renameTarget = $state<{ kind: 'file' | 'image'; index: number } | null>(null);
	let renameValue = $state('');
	let renameError = $state('');
	let renameTitle = $derived(renameTarget?.kind === 'file' ? 'Rename File' : 'Rename Image');

	// delete (shared for files + images)
	let deleteTarget = $state<{ kind: 'file' | 'image'; index: number } | null>(null);
	let deleteTitle = $derived(deleteTarget?.kind === 'file' ? 'Delete File' : 'Delete Image');
	let deleteName = $derived(
		deleteTarget
			? deleteTarget.kind === 'file'
				? projectData.codeFiles[deleteTarget.index]?.name
				: projectData.images[deleteTarget.index]?.name
			: ''
	);

	function openNewImageDialog() {
		newImageData = { name: '', width: 0, height: 0 };
		newImageError = '';
		newImageDialogOpen = true;
	}

	function openNewFileDialog() {
		newFileName = '';
		newFileError = '';
		newFileDialogOpen = true;
	}

	function openRename(kind: 'file' | 'image', index: number) {
		renameTarget = { kind, index };
		renameValue =
			kind === 'file' ? projectData.codeFiles[index].name : projectData.images[index].name;
		renameError = '';
		renameDialogOpen = true;
	}

	function openDelete(kind: 'file' | 'image', index: number) {
		deleteTarget = { kind, index };
		deleteDialogOpen = true;
	}

	function createNewImage(event: Event) {
		event.preventDefault();
		const { name, width, height } = newImageData;
		if (
			!imageNameIsValid(name.trim()) ||
			!Number.isInteger(width) ||
			!Number.isInteger(height) ||
			width <= 0 ||
			height <= 0
		) {
			newImageError = 'Please fill in all fields with valid values.';
			return;
		}
		if (imageNameExists(name.trim())) {
			newImageError = 'An image with this name already exists.';
			return;
		}
		if (width > MAX_IMAGE_DIM || height > MAX_IMAGE_DIM) {
			newImageError = `Dimensions must be ${MAX_IMAGE_DIM}\u00d7${MAX_IMAGE_DIM} or less.`;
			return;
		}
		let newImageContent = new Uint8Array(width * height * 2); //initialise it blank so it doens't crash the device when we try to render

		projectData.images.push({
			name: name.trim(),
			width,
			height,
			content: newImageContent.toBase64()
		});
		newImageData = { name: '', width: 0, height: 0 };
		newImageError = '';
		newImageDialogOpen = false;
	}

	function createNewFile(event: Event) {
		event.preventDefault();
		const name = newFileName.trim();
		if (!name) {
			newFileError = 'Please enter a file name.';
			return;
		}
		if (fileNameError(name)) {
			newFileError = fileNameError(name);
			return;
		}
		if (name === MAIN_FILE) {
			newFileError = `${MAIN_FILE} is reserved.`;
			return;
		}
		if (fileNameExists(name)) {
			newFileError = 'A file with this name already exists.';
			return;
		}
		projectData.codeFiles.push({ name, content: 'import ferret\n\n# Your code here!\n' });
		newFileName = '';
		newFileError = '';
		newFileDialogOpen = false;
		currentFileIndex = projectData.codeFiles.length - 1;
		editor?.loadNewEditorContent('import ferret\n\n# Your code here!\n');
	}

	function confirmRename(event: Event) {
		event.preventDefault();
		if (!renameTarget) return;
		const name = renameValue.trim();
		if (!name) {
			renameError = 'Name cannot be empty.';
			return;
		}
		if (renameTarget.kind === 'file') {
			if (fileNameError(name)) {
				renameError = fileNameError(name);
				return;
			}
			if (name === MAIN_FILE) {
				renameError = `${MAIN_FILE} is reserved and cannot be used.`;
				return;
			}
			if (fileNameExists(name, renameTarget.index)) {
				renameError = 'A file with this name already exists.';
				return;
			}
			projectData.codeFiles[renameTarget.index].name = name;
		} else {
			if (!imageNameIsValid(name)) {
				renameError = 'Image names cannot contain /, \\, :, commas, or new lines.';
				return;
			}
			if (imageNameExists(name, renameTarget.index)) {
				renameError = 'An image with this name already exists.';
				return;
			}
			projectData.images[renameTarget.index].name = name;
		}
		renameDialogOpen = false;
		renameTarget = null;
	}

	function confirmDelete() {
		if (!deleteTarget) return;
		if (deleteTarget.kind === 'file') {
			const deletedIndex = deleteTarget.index;
			projectData.codeFiles.splice(deletedIndex, 1);
			if (projectData.codeFiles.length === 0) {
				currentFileIndex = 0;
			} else if (deletedIndex === currentFileIndex) {
				currentFileIndex = Math.min(currentFileIndex, projectData.codeFiles.length - 1);
			} else if (deletedIndex < currentFileIndex) {
				currentFileIndex -= 1;
			}
			editor?.loadNewEditorContent(projectData.codeFiles[currentFileIndex]?.content ?? '');
		} else {
			projectData.images.splice(deleteTarget.index, 1);
			if (selectedImageIndex === deleteTarget.index) {
				selectedImageIndex = null;
			} else if (selectedImageIndex !== null && selectedImageIndex > deleteTarget.index) {
				selectedImageIndex -= 1;
			}
		}
		deleteDialogOpen = false;
		deleteTarget = null;
	}

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

	function downloadProject() {
		const dataStr =
			'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(projectData));
		const downloadAnchorNode = document.createElement('a');
		downloadAnchorNode.setAttribute('href', dataStr);
		downloadAnchorNode.setAttribute('download', projectData.name + '.json');
		document.body.appendChild(downloadAnchorNode); // required for firefox
		downloadAnchorNode.click();
		downloadAnchorNode.remove();
	}

	async function saveToDevice() {
		if (!connection || !connection.connected) {
			alert('connect before sending.');
			return;
		}
		try {
			await connection.saveProjectToDevice(projectData);
		} catch (err) {
			console.error('device save failed', err);
			alert(`Device save failed: ${err instanceof Error ? err.message : String(err)}`);
		}
	}

	function loadProject() {
		const savedProjectData = localStorage.getItem('projectData');
		if (savedProjectData) {
			projectData = JSON.parse(savedProjectData);
		}
		// The editor only reads its bound content once on mount, so after
		// restoring the project we must explicitly reload the current file.
		editor?.loadNewEditorContent(projectData.codeFiles[currentFileIndex].content);
	}

	function loadProjectFromFile(event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			const file = input.files[0];
			const reader = new FileReader();
			reader.onload = (e) => {
				const content = e.target?.result as string;
				projectData = JSON.parse(content);
				editor?.loadNewEditorContent(projectData.codeFiles[currentFileIndex].content);
			};
			reader.readAsText(file);
		}
	}

	function handleSubmit(event: Event) {
		event.preventDefault();
		sendTerminalInput();
	}

	function save() {
		localStorage.setItem('projectData', JSON.stringify(projectData));
	}

	let connection = $state<WebSerialConnection | null>(null);
	onMount(() => {
		loadProject();
		connection = getWebSerialConnection();
		void connection.init();
	});

	async function uploadScript() {
		if (!connection || !connection.connected) {
			alert('Not connected to the device.');
			return;
		}
		try {
			await connection.runProject(projectData);
		} catch (err) {
			console.error('project upload failed', err);
			alert(`Project upload failed: ${err instanceof Error ? err.message : String(err)}`);
		}
	}
</script>

{#if connection}
	<div class="flex h-full w-full flex-col overscroll-none">
		<div class="flex w-full flex-row">
			<Button onclick={connection.connect} class="h-full">Connect</Button>
			{#if connection.connected}
				<Button variant="secondary" onclick={connection.controlC} class="h-full"
					>Send control c</Button
				>
				<Button variant="secondary" onclick={connection.controlD} class="h-full"
					>Send control d</Button
				>
				<Button variant="secondary" onclick={connection.controlB} class="h-full"
					>Send control B</Button
				>
				<Button variant="secondary" onclick={connection.disconnect} class="h-full"
					>Disconnect</Button
				>
				<Button onclick={uploadScript} class="h-full">Run Script</Button>
				<Button onclick={saveToDevice} class="h-full">Save to device</Button>
				<Button onclick={syncImages} class="h-full">Sync images</Button>
			{/if}
			<Button variant="secondary" onclick={save} class="h-full"
				>Save project to local storage</Button
			>
			<Button variant="secondary" onclick={downloadProject} class="h-full"
				>Download project to file</Button
			>
			<Button variant="secondary" class="h-full">
				<label for="file-upload" class="cursor-pointer">Load project from file</label>
				<input
					id="file-upload"
					type="file"
					accept=".json"
					onchange={loadProjectFromFile}
					class="hidden"
				/>
			</Button>
			<div class="ml-auto flex items-center gap-2 p-2">
				<ThemeToggle />
			</div>
		</div>
		<Resizable.PaneGroup
			direction="horizontal"
			class="min-h-0 max-w-screen flex-1 rounded-lg border"
		>
			<Resizable.Pane defaultSize={70} minSize={30}>
				<Resizable.PaneGroup
					direction="vertical"
					class="min-h-50 max-w-md min-w-screen rounded-lg border"
				>
					<Resizable.Pane defaultSize={70} minSize={30}>
						<Editor
							bind:editorContent={projectData.codeFiles[currentFileIndex].content}
							bind:this={editor}
						></Editor>
					</Resizable.Pane>
					<Resizable.Handle withHandle />
					<Resizable.Pane defaultSize={30} minSize={10}>
						<div class="flex h-full min-h-0 w-full flex-col">
							{#if connection.connected}
								<pre
									class="min-h-0 flex-1 overflow-y-auto whitespace-pre-wrap"
									bind:this={outputelement}>{connection.output}</pre>
							{:else}
								<p class="min-h-0 flex-1 overflow-y-auto">
									Connect to the device to see output and send terminal commands.
								</p>
							{/if}
							<form onsubmit={handleSubmit} class="flex w-full flex-row justify-between">
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
			</Resizable.Pane>
			<ResizableHandle withHandle />
			<Resizable.Pane defaultSize={30} minSize={10}>
				<Accordion.Root
					type="multiple"
					value={['files', 'images']}
					class="flex h-full w-full flex-col"
				>
					<Accordion.Item value="files" class="flex flex-col">
						<Accordion.Trigger class="px-4">Files</Accordion.Trigger>
						<Accordion.Content class="flex-1 overflow-y-auto">
							<div class="flex flex-col gap-1 p-2">
								<Button
									variant="outline"
									size="xs"
									class="w-full justify-center"
									onclick={openNewFileDialog}
								>
									<FilePlus /> New File
								</Button>
								<div class="flex flex-col">
									{#each projectData.codeFiles as file, index (file.name)}
										<div
											class="group flex items-center gap-1 px-1.5 py-1.5 {index === currentFileIndex
												? 'bg-accent text-accent-foreground'
												: 'hover:bg-muted/60'}"
										>
											<button
												class="flex min-w-0 flex-1 items-center gap-1.5 text-left text-xs"
												onclick={() => changeActiveFile(index)}
											>
												<FileCode
													class="size-3.5 shrink-0 {file.name === 'main.py'
														? 'text-primary'
														: 'text-muted-foreground'}"
												/>
												<span class="truncate">{file.name}</span>
											</button>
											{#if file.name !== 'main.py'}
												<div class="flex items-center gap-0.5">
													<Button
														size="icon-xs"
														variant="ghost"
														title="Rename"
														onclick={() => openRename('file', index)}
													>
														<PencilLine />
													</Button>
													<Button
														size="icon-xs"
														variant="ghost"
														title="Delete"
														class="hover:text-destructive"
														onclick={() => openDelete('file', index)}
													>
														<Trash />
													</Button>
												</div>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						</Accordion.Content>
					</Accordion.Item>
					<!-- <Accordion.Item value="docs" class="flex flex-col">
						<Accordion.Trigger class="px-4">Docs</Accordion.Trigger>
						<Accordion.Content class="px-4 flex-1">
							<SidebarDocs />
						</Accordion.Content>
					</Accordion.Item> -->
					<Accordion.Item value="images" class="flex flex-col">
						<Accordion.Trigger class="px-4">Images</Accordion.Trigger>
						<Accordion.Content class="flex-1 overflow-y-auto px-2">
							<div class="flex flex-col gap-1">
								<Button
									variant="outline"
									size="xs"
									class="w-full justify-center"
									onclick={openNewImageDialog}
								>
									<ImagePlus /> New Image
								</Button>
								{#if projectData.images.length === 0}
									<p class="px-2 py-4 text-center text-xs text-muted-foreground">No images yet</p>
								{:else}
									<div class="flex flex-col">
										{#each projectData.images as image, index (image.name)}
											<div
												class="group flex items-center gap-1 px-1.5 py-1.5 {index ===
												selectedImageIndex
													? 'bg-accent text-accent-foreground'
													: 'hover:bg-muted/60'}"
											>
												<button
													class="flex min-w-0 flex-1 items-center gap-1.5 text-left text-xs"
													onclick={() => (selectedImageIndex = index)}
												>
													<ImageIcon class="size-3.5 shrink-0 text-muted-foreground" />
													<span class="truncate">{image.name}</span>
													<span
														class="ml-auto shrink-0 text-[10px] text-muted-foreground tabular-nums"
													>
														{image.width}×{image.height}
													</span>
												</button>
												<div class="flex items-center gap-0.5">
													<Button
														size="icon-xs"
														variant="ghost"
														title="Rename"
														onclick={() => openRename('image', index)}
													>
														<PencilLine />
													</Button>
													<Button
														size="icon-xs"
														variant="ghost"
														title="Delete"
														class="hover:text-destructive"
														onclick={() => openDelete('image', index)}
													>
														<Trash />
													</Button>
												</div>
											</div>
										{/each}
									</div>
								{/if}
								{#if selectedImageIndex !== null}
									<div class="mt-1 border-t border-border">
										<ImageEditor bind:imageData={projectData.images[selectedImageIndex]} />
									</div>
								{/if}
							</div>
						</Accordion.Content>
					</Accordion.Item>
					<div class="flex-1 border-0"></div>
					<!-- note: this bottom div is just to take up all the remaining space -->
				</Accordion.Root>
			</Resizable.Pane>
		</Resizable.PaneGroup>

		<!-- New Image dialog -->
		<Dialog.Root bind:open={newImageDialogOpen}>
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>New Image</Dialog.Title>
					<Dialog.Description>
						Name must be unique. Dimensions up to {MAX_IMAGE_DIM}×{MAX_IMAGE_DIM}.
					</Dialog.Description>
				</Dialog.Header>
				<form onsubmit={createNewImage} class="flex flex-col gap-2">
					<Input bind:value={newImageData.name} placeholder="Image Name" />
					<div class="flex gap-2">
						<Input
							type="number"
							bind:value={newImageData.width}
							placeholder="Width"
							min="1"
							max={MAX_IMAGE_DIM}
						/>
						<Input
							type="number"
							bind:value={newImageData.height}
							placeholder="Height"
							min="1"
							max={MAX_IMAGE_DIM}
						/>
					</div>
					{#if newImageError}
						<p class="flex items-center gap-1.5 text-xs text-destructive">
							<TriangleAlert class="size-3.5 shrink-0" />
							{newImageError}
						</p>
					{/if}
					<Dialog.Footer>
						<Button variant="outline" onclick={() => (newImageDialogOpen = false)}>Cancel</Button>
						<Button type="submit">Create</Button>
					</Dialog.Footer>
				</form>
			</Dialog.Content>
		</Dialog.Root>

		<!-- New File dialog -->
		<Dialog.Root bind:open={newFileDialogOpen}>
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>New File</Dialog.Title>
					<Dialog.Description>Name must be unique and cannot be main.py.</Dialog.Description>
				</Dialog.Header>
				<form onsubmit={createNewFile} class="flex flex-col gap-2">
					<Input bind:value={newFileName} placeholder="file.py" />
					{#if newFileError}
						<p class="flex items-center gap-1.5 text-xs text-destructive">
							<TriangleAlert class="size-3.5 shrink-0" />
							{newFileError}
						</p>
					{/if}
					<Dialog.Footer>
						<Button variant="outline" onclick={() => (newFileDialogOpen = false)}>Cancel</Button>
						<Button type="submit">Create</Button>
					</Dialog.Footer>
				</form>
			</Dialog.Content>
		</Dialog.Root>

		<!-- Rename dialog (files + images) -->
		<Dialog.Root bind:open={renameDialogOpen}>
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>{renameTitle}</Dialog.Title>
				</Dialog.Header>
				<form onsubmit={confirmRename} class="flex flex-col gap-2">
					<Input bind:value={renameValue} placeholder="Name" />
					{#if renameError}
						<p class="flex items-center gap-1.5 text-xs text-destructive">
							<TriangleAlert class="size-3.5 shrink-0" />
							{renameError}
						</p>
					{/if}
					<Dialog.Footer>
						<Button variant="outline" onclick={() => (renameDialogOpen = false)}>Cancel</Button>
						<Button type="submit">Rename</Button>
					</Dialog.Footer>
				</form>
			</Dialog.Content>
		</Dialog.Root>

		<!-- Delete dialog (files + images) -->
		<Dialog.Root bind:open={deleteDialogOpen}>
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>{deleteTitle}</Dialog.Title>
					<Dialog.Description>
						Are you sure you want to delete <span class="font-medium text-foreground"
							>{deleteName}</span
						>? This cannot be undone.
					</Dialog.Description>
				</Dialog.Header>
				<Dialog.Footer>
					<Button variant="outline" onclick={() => (deleteDialogOpen = false)}>Cancel</Button>
					<Button variant="destructive" onclick={confirmDelete}>Delete</Button>
				</Dialog.Footer>
			</Dialog.Content>
		</Dialog.Root>
	</div>
{/if}
