<script>
  import { marked } from "marked";
  import StarterKit from "@tiptap/starter-kit";
  import { Editor } from "@tiptap/core";
  import TaskItem from "@tiptap/extension-task-item";
  import TaskList from "@tiptap/extension-task-list";
  import BulletList from '@tiptap/extension-bullet-list'
  import { onMount } from "svelte";
  import "./tailwind.css"
  let notes = fetch("/api/notes").then((res) => res.json());

  marked.setOptions({
    renderer: new marked.Renderer(),
    gfm: true,
    breaks: true,
    sanitize: false,
  });

  let context = "work";
  let selectedNoteId = null;
  let editingNoteId = null;
  let editingNoteValue = "";
  let editingNoteTags = "";
  let editorValue = "";
  let editorTags = context + ", ";
  let element;
  let editor;

  function handleMouseOver(noteId) {
    selectedNoteId = noteId;
  }

  function handleNoteDelete(noteId) {
    fetch(`/api/note/${noteId}`, { method: "delete" });
    notes = notes.then((p) => p.filter((note) => note.id !== noteId));
  }

  function handleNoteEdit(noteId) {
    editingNoteId = noteId;
    notes.then((p) => {
      const note = p.find((note) => note.id === noteId);
      editingNoteValue = note.body;
      editingNoteTags = note.tags.join(", ");
    });
  }

  function handleEditingSubmit(e) {
    e.preventDefault();
    let note = { body: editingNoteValue, tags: editingNoteTags.split(", ") };
    fetch(`/api/note/${editingNoteId}`, {
      method: "put",
      body: JSON.stringify(note),
    });
    notes.then((p) => {
      const noteIndex = p.findIndex((note) => note.id !== editingNoteId);
      note.id = editingNoteId;
      notes = [
        ...p.slice(0, noteIndex),
        note,
        ...p.slice(noteIndex + 1, p.length),
      ];
    });

    editingNoteId = null;
  }

  function handleMouseOut() {
    selectedNoteId = null;
  }

  function handleEditorSubmit(e) {
    e.preventDefault();
    let note = { body: editorValue, tags: editorTags.split(", ") };
    fetch("/api/note", { method: "post", body: JSON.stringify(note) });
    notes = [note, ...notes];
  }

  onMount(() => {
    editor = new Editor({
      element: element,
      extensions: [
        StarterKit,
        TaskItem,
        TaskList.configure({ HTMLAttributes: { class: "not-prose" } }),
        BulletList.configure({HTMLAttributes: {class: "not-prose"}})
      ],
      autofocus: true,
      editorProps: {
        attributes: {
          class:
            "prose prose-sm sm:prose lg:prose-lg xl:prose-2xl m-5 focus:outline-none border rounded",
        },
      },
      content: "",
      onTransaction: () => {
        // force re-render so `editor.isActive` works as expected
        editor = editor;
      },
    });
  });
</script>

<main class="mx-auto max-w-lg">
  <header>
    <div class="flex justify-between items-baseline">
      <h1 class="prose text-6xl">
        <a href="/">Nous</a>
        <sup hx-trigger="load" hx-get="/active-context" hx-swap="innerHTML" />
      </h1>
      <form class="search-form" action="/search" method="get">
        <input type="text" name="query" required placeholder="search" />
        <input type="submit" value="Search" />
      </form>
    </div>
    <nav>
      <a href="/todos">Todos</a>
      <a href="/tag?tags=to read">Readings</a>
      <a href="/review">Review</a>
    </nav>
  </header>
  <form class="m-2">
    <div bind:this={element} />
    <input
      class="rounded block"
      type="text"
      name="tags"
      placeholder="use comma 'seperated values'"
      autocorrect="off"
      autocapitalize="none"
      bind:value={editorTags}
    />
    <input
      class="bg-gray-200 rounded p-1 block my-2"
      type="submit"
      value="Submit"
      on:click={handleEditorSubmit}
    />
  </form>
  <hr class="my-5" />
  {#await notes}
    <p>Loading notes</p>
  {:then value}
    {#each value as note}
      <div
        on:mouseover={() => handleMouseOver(note.id)}
        on:focus={() => handleMouseOver(note.id)}
        on:blur={handleMouseOut}
        on:mouseout={handleMouseOut}
      >
        <div class="flex justify-end space-x-1">
          <div
            class="cursor-pointer rounded bg-yellow-600 p-1 inline-block"
            on:click={() => handleNoteEdit(note.id)}
          >
            Edit
          </div>
          <div
            class="cursor-pointer rounded bg-red-800 p-1 inline-block"
            on:click={() => handleNoteDelete(note.id)}
          >
            Delete
          </div>
        </div>
        <div
          class="prose text-white bg-gray-400 my-2 rounded p-2 group-over:shadow transition z-20 relative"
        >
          {#if editingNoteId === note.id}
            <form class="m-2 text-black">
              <textarea
                class="rounded min-w-full"
                type="text"
                name="body"
                required
                autofocus
                bind:value={editingNoteValue}
              />
              <input
                class="rounded block"
                type="text"
                name="tags"
                placeholder="use comma 'seperated values'"
                autocorrect="off"
                autocapitalize="none"
                bind:value={editingNoteTags}
              />
              <input
                class="bg-gray-200 rounded p-1 block my-2"
                type="submit"
                value="Submit"
                on:click={handleEditingSubmit}
              />
            </form>
          {:else}
            {@html marked.parse(note.body)}
          {/if}
        </div>
        {#each note.tags as tag}
          <div
            class="p-1 mr-1 bg-gray-400 rounded inline-block text-sm relative"
          >
            {tag}
          </div>
        {/each}
      </div>
    {/each}
  {:catch error}
    <p>Something went wrong: {error.message}</p>
  {/await}
</main>
