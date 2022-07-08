<script>
  import { marked } from "marked";

  marked.setOptions({
    gfm: true,
    breaks: true,
    smartLists: true,
});

  export let previousDay;
  export let nextDay;
  export let currentDay;
  export let notes;
  let editingId = undefined;
  let editingBody = "";
  let editingTags = "";

  async function postData(url = '', data = {}, method = "POST") {
    // Default options are marked with *
    const response = await fetch(url, {
      method: method,
      mode: 'cors',
      cache: 'no-cache',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
      redirect: 'follow',
      referrerPolicy: 'no-referrer',
      body: JSON.stringify(data)
    })
      .then(response => {
        if (!response.ok) {
          throw new Error('Network response was not OK');
        }
        return response;
        })
      .catch(error => {
        alert("there was an error", error)
        throw new Error('Network response was not OK');
      })
  }

  function toggle(id) {
    let success = false;
    postData('/api/done', { Id: id })
    .then(() => {
      notes = notes.map(note => {
        if (note.ID === id) {
          // return a new object
          return {
            ID: id,
            Done: !note.Done,
            BodyRaw: note.BodyRaw,
            Tags: note.Tags
          };
        }

        // return the same object
        return note;
		});

    })
    .catch(() => {
        alert('there as a boo boo')
    })
  }

  function toggleEdit(id) {
    if (id === editingId) {
      editingId = undefined
      return;
    }
    const note = notes.find(note => {
      if (note.ID === id) {
        return note
      }
    });
    editingId = id
    editingBody = note.BodyRaw
    editingTags = note.Tags.join(", ")
  }

  function handleEdit(id) {
    postData(`/api/note/${id}`, { Id: id, Body: editingBody, Tags: editingTags  }, "PUT")
    .then(response => {
      notes = notes.map(note => {
        if (note.ID === id) {
          // return a new object
          return {
            ID: id,
            Done: note.Done,
            BodyRaw: editingBody,
            Tags: editingTags.split(", ")
          };
        }

        // return the same object
        return note;
		}
        )
          editingId = undefined
      });
  }

</script>

<style>
  .emoji-button {
    cursor: pointer;
  }

</style>
<main>
  <div class="prev-next">
    <a href={`/t/${previousDay}`}>&larr;</a>
    <h2>{currentDay}</h2>
    <a href={`/t/${nextDay}`}>&rarr;</a>
  </div>
  <form class="submit" action="/note" method="post">
    <textarea type="text" name="body" required autofocus />
    <input type="text" name="tags" placeholder="use comma 'seperated values'" autocorrect="off" autocapitalize="none" />
    <input type="submit" value="Submit" />
  </form>
  <div class="grid-note">
    {#each notes as note}
      <div class="controls">
        <input
          name="done"
          type="checkbox"
          checked={note.Done}
          on:change={() => toggle(note.ID)}
        />
        <div class="emoji-button" on:click={()=>toggleEdit(note.ID)}>
          {#if editingId === note.ID}
          &#10060;
          {:else}
          &#128397;
          {/if}
        </div>
        <div class="emoji-button" on:click={() => location.href=`/note/${note.ID}/delete`}>&#x1F5D1;</div>
      </div>
    {#if editingId === note.ID}
      <div class="note submit" >
        <textarea type="text" name="body" required bind:value={editingBody}/>
        <input type="text" name="tags" placeholder="use comma 'seperated values'" bind:value={editingTags} autocorrect="off" autocapitalize="none"/>
        <input type="submit" value="Submit" on:click={()=>handleEdit(note.ID)}/>
      </div>
    {:else}
        <div class="note" >
      <div class:done={note.Done}>
        <a name={note.ID} />
        {@html marked(note.BodyRaw)}
      </div>
      <div class="metadata">
        <ul class="tags text-subdued">
          {#each note.Tags as tag}
            <li class="tag">
                  <a href={`/search?query=${tag}`}>{tag}</a>
                </li>
          {/each}
        </ul>
      </div>
        <hr />
        </div>
    {/if}
    {/each}
  </div>
</main>
