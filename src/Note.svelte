<script lang="ts">
  import Editor, { renderHTML } from "./editor.svelte";
  import { MarkdownParser } from "prosemirror-markdown";
  import { schema } from "./schema";
  export let deleteCallback: any;
  export let note: any;
  import markdownit from 'markdown-it'

  let editing = false;

function listIsTight(tokens: readonly any[], i: number) {
  while (++i < tokens.length)
    if (tokens[i].type != "list_item_open") return tokens[i].hidden
  return false
}

  function handleDelete(noteId: any) {
    fetch(`/api/note/${noteId}`, { method: "delete" }).then(deleteCallback);
  }

  const mp = new MarkdownParser(
    schema,
    markdownit("commonmark", { html: false }),
    {
      blockquote: { block: "blockquote" },
      paragraph: { block: "paragraph" },
      list_item: { block: "listItem" },
      bullet_list: {
        block: "bulletList",
        getAttrs: (_, tokens, i) => ({ tight: listIsTight(tokens, i) }),
      },
      ordered_list: {
        block: "orderedList",
        getAttrs: (tok, tokens, i) => ({
          order: +tok.attrGet("start") || 1,
          tight: listIsTight(tokens, i),
        }),
      },
      heading: {
        block: "heading",
        getAttrs: (tok) => ({ level: +tok.tag.slice(1) }),
      },
      code_block: { block: "codeBlock", noCloseToken: true },
      fence: {
        block: "codeBlock",
        getAttrs: (tok) => ({ params: tok.info || "" }),
        noCloseToken: true,
      },
      hr: { node: "horizontalRule" },
      image: {
        node: "image",
        getAttrs: (tok) => ({
          src: tok.attrGet("src"),
          title: tok.attrGet("title") || null,
          alt: (tok.children[0] && tok.children[0].content) || null,
        }),
      },
      hardbreak: { node: "hardBreak" },

      em: { mark: "italic" },
      strong: { mark: "bold" },
      link: {
        mark: "link",
        getAttrs: (tok) => ({
          href: tok.attrGet("href"),
          title: tok.attrGet("title") || null,
        }),
      },
      code_inline: { mark: "code", noCloseToken: true },
    }
  );

  function foo() {
    try {
      return renderHTML(JSON.stringify(mp.parse(note.body)));
    } catch (error) {
      console.log(note.body);
      console.error(error);
    }
  }
</script>

<div
  class="prose my-2 rounded p-2 group-over:shadow transition z-20 relative border"
>
  {#if editing}
    <!--TODO: We probably need to use a store or something to trigger a refresh -->
    <Editor {note} submitCallback={deleteCallback} context={""} />
  {:else}
    {@html foo()}
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
  <button
    class="rounded border p-1 bg-yellow-100"
    on:click={() => (editing = true)}
  >
    Edit
  </button>
</div>
