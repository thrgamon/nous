<script>
  import Editor, { renderHTML } from "./editor.svelte";
  export let deleteCallback;
  export let note;

  let editing = false;

  function handleDelete(noteId) {
    fetch(`/api/note/${noteId}`, { method: "delete" }).then(deleteCallback);
  }
</script>

<div
  class="prose my-2 rounded p-2 group-over:shadow transition z-20 relative border"
>
  {#if editing}
    <!--TODO: We probably need to use a store or something to trigger a refresh -->
    <Editor {note} submitCallback={deleteCallback}/>
  {:else}
    {@html renderHTML(note.body)}
  {/if}
</div>
<div class="flex justify-between">
  <div>
    {#each note.tags as tag}
      <div class="p-1 mr-1 rounded inline-block text-sm relative border">
        {tag}
      </div>
    {/each}
  </div>
  <button
    class="rounded border p-1 bg-red-100"
    on:click={() => handleDelete(note.id)}
  >
    Delete
  </button>
  <button class="rounded border p-1 bg-yellow-100" on:click={() => editing = true}>
    Edit
  </button>
</div>
