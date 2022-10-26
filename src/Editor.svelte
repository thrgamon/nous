<script context="module">
  import extensions from "./extensions.js"

  export function renderHTML(content) {
    return generateHTML(JSON.parse(content), extensions);
  }
</script>

<script>
  export let note;
  export let context;
  export let submitCallback;

  import { Editor, generateHTML } from "@tiptap/core";
  import { onMount } from "svelte";

  let element;
  let editor;
  let body = "";
  let tags = context + ", ";
  let editing = false;
  if (note) {
    body = JSON.parse(note.body);
    tags = note.tags.join(", ") + ", ";
    editing = true;
  }

  function createNote() {
    let newNote = {
      body: JSON.stringify(editor.getJSON()),
      tags: formatTags(tags),
    };
    fetch("/api/note", { method: "post", body: JSON.stringify(newNote) })
      .then(editor.commands.clearContent(true))
      .then(submitCallback);
  }

  function editNote() {
    let newNote = {
      body: JSON.stringify(editor.getJSON()),
      tags: formatTags(tags),
    };
    fetch(`/api/note/${note.id}`, {
      method: "put",
      body: JSON.stringify(newNote),
    }).then(submitCallback);
  }

  function handleSubmit() {
    if (editing) {
      editNote(note.id);
    } else {
      createNote();
    }
  }

  function formatTags(tagString) {
    tagString = tagString.trim();

    if (tagString == "") {
      return [];
    }

    if (tagString.endsWith(",")) {
      tagString = tagString.slice(0, -1);
    }

    return tagString.split(",");
  }

  function initialiseEditor() {
    return new Editor({
      element: element,
      extensions: extensions,
      autofocus: true,
      editorProps: {
        attributes: {
          class:
            "prose prose-sm sm:prose lg:prose-lg xl:prose-2xl m-5 focus:outline-none border rounded p-0",
        },
      },
      content: body,
    });
  }

  onMount(() => {
    editor = initialiseEditor();
  });
</script>

<form
  class="m-2"
  on:submit|preventDefault={handleSubmit}
  on:keydown={(e) => {
    if (e.metaKey && e.key == "Enter") {
      handleSubmit();
    }
  }}
>
  <div bind:this={element} />
  <input
    class="rounded block"
    type="text"
    name="tags"
    placeholder="use comma 'seperated values'"
    autocorrect="off"
    autocapitalize="none"
    bind:value={tags}
  />
  <input
    class="bg-gray-200 rounded p-1 block my-2 cursor-pointer"
    value="Submit"
    type="submit"
  />
</form>
