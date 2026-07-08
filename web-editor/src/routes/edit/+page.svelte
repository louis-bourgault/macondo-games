<script lang="ts">
    import Editor from '$lib/components/ui/editor.svelte';
    let editorContent = $state(`from modules.display import Display`)
</script>

<Editor bind:editorContent></Editor>