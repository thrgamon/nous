<script>
  import "./tailwind.css";
  import Editor from "./editor.svelte";
  import Note from "./note.svelte";

  let context = "work";
  var notes = [];

  fetchNotes();

  function fetchNotes() {
    fetch("/api/notes?from=2022-01-25&to=2022-10-01").then(
      (res) => (notes = res.json())
    );
  }

</script>

<main class="mx-auto max-w-lg">
  <header>
    <div class="flex justify-between items-baseline">
      <h1 class="prose text-6xl">
        <a href="/">Nous</a>
      </h1>
    </div>
  </header>
  <Editor context="work" submitCallback={fetchNotes} />
  <hr class="my-5" />
  {#await notes}
    <p>Loading notes</p>
  {:then value}
    {#each value as note}
      <Note {note} deleteCallback={fetchNotes} />
    {/each}
  {:catch error}
    <p>Something went wrong: {error.message}</p>
  {/await}
</main>
