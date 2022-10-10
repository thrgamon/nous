<script>
  import { marked } from "marked";
  let notes = fetch("/api/notes").then((res) => res.json());

  marked.setOptions({
    renderer: new marked.Renderer(),
    gfm: true,
    breaks: true,
    sanitize: false,
  });

  let context = "work";
  let blur = false;
  let selectedNoteId = null;
  let editorValue = "";
  let editorTags = context + ", ";

  function handleMouseOver(noteId) {
    selectedNoteId = noteId;
    blur = true;
  }

  function handleNoteDelete(noteId) {
    fetch(`/api/note/${noteId}`, { method: "delete" });
    notes = notes.then((p) => p.filter((note) => note.id !== noteId));
  }

  function handleMouseOut() {
    selectedNoteId = null;
    blur = false;
  }

  function handleEditorSubmit(e) {
    let note = { body: editorValue, tags: ["hello"] };
    fetch("/api/note", { method: "post", body: JSON.stringify(note) });
    notes = [note, ...notes];
    e.preventDefault();
  }
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
  <form
    class="m-2 g"
    hx-post="/note"
    hx-trigger="submit, keydown[metaKey&&(keyCode==10||keyCode==13)]"
  >
    <textarea
      class="rounded min-w-full"
      type="text"
      name="body"
      required
      autofocus
      bind:value={editorValue}
    />
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
        class="peer group hover:scale-105 transition hover:z-30 relative"
        class:blur-lg={blur && note.id !== selectedNoteId}
        on:mouseover={() => handleMouseOver(note.id)}
        on:focus={() => handleMouseOver(note.id)}
        on:blur={handleMouseOut}
        on:mouseout={handleMouseOut}
      >
        <div
          class="prose text-gray-50 bg-slate-400 my-2 rounded p-2 group-over:shadow transition z-20 relative"
        >
          {@html marked.parse(note.body)}
        </div>
        {#each note.tags as tag}
          <div
            class="p-0.5 mr-1 bg-gray-600 rounded -translate-y-10 group-hover:translate-y-0 transition z-10 inline-block text-sm relative"
          >
            {tag}
          </div>
        {/each}
        <div on:click={() => handleNoteDelete(note.id)}>Delete</div>
      </div>
    {/each}
  {:catch error}
    <p>Something went wrong: {error.message}</p>
  {/await}
</main>
