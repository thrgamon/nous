<script>
  import { marked } from "marked";

  export let previousDay;
  export let nextDay;
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
    .then(response => {
      notes = notes.map(note => {
        if (note.ID === id) {
          // return a new object
          return {
            ID: id,
            Done: !note.Done,
            Body: note.Body,
            Tags: note.Tags
          };
        }

        // return the same object
        return note;
		});

      console.log(notes)
    })
    .catch(error => {
        alert('there as a boo boo')
    })
  }

  function toggleEdit(id) {
      const note = notes.find(note => {
        if (note.ID === id) {
          // return a new object
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

<main>
  <div class="prev-next">
    <a href={`/t/${previousDay}`}>&larr;</a>
    {#if nextDay}
      <a href={`/t/${nextDay}`}>&rarr;</a>
    {/if}
  </div>
  <form class="submit" action="/note" method="post">
    <textarea type="text" name="body" required autofocus />
    <input type="text" name="tags" placeholder="use comma 'seperated values'" />
    <input type="submit" value="Submit" />
  </form>
  <div class="grid-note">
    {#each notes as note}
    {#if editingId === note.ID}
      <textarea type="text" name="body" required bind:value={editingBody}/>
      <input type="text" name="tags" placeholder="use comma 'seperated values'" bind:value={editingTags}/>
      <input type="submit" value="Submit" on:click={()=>handleEdit(note.ID)}/>
      <button on:click={()=>toggleEdit()}>X</button>
    {:else}
        <div class="note" >
      <div class:done={note.Done} on:click={() => toggleEdit(note.ID)}>
        <a name={note.ID} />
        {@html marked(note.BodyRaw)}
      </div>
      <div class="metadata" on:click={() => toggleEdit(note.ID)}>
        <ul class="tags text-subdued">
          {#each note.Tags as tag}
            <li class="tag">{tag}</li>
          {/each}
        </ul>
      </div>
        <hr />
        </div>
    {/if}
      <div class="controls">
        <input
          name="done"
          type="checkbox"
          checked={note.Done}
          on:change={() => toggle(note.ID)}
        />
        <a href="/note/{note.ID}/delete">delete</a>
      </div>
    {/each}
  </div>
</main>
