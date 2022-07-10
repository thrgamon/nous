<script>
  import { marked } from "marked";
  import Metadata from "./Metadata.svelte"

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
  let toggleDone = true;

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
        if (note.id === id) {
          // return a new object
          return {
            id: id,
            done: !note.done,
            body: note.body,
            tags: note.tags
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
      if (note.id === id) {
        return note
      }
    });
    editingId = id
    editingBody = note.body
    editingTags = note.tags.join(", ")
  }

  function handleEdit(id) {
    postData(`/api/note/${id}`, { Id: id, Body: editingBody, tags: editingTags  }, "PUT")
    .then(response => {
      notes = notes.map(note => {
        if (note.id === id) {
          // return a new object
          return {
            id: id,
            done: note.done,
            body: editingBody,
            tags: editingTags.split(", ")
          };
        }

        // return the same object
        return note;
		}
        )
          editingId = undefined
      });
  }

  function toggleDoneFilter() {
    toggleDone = !toggleDone
    document.querySelectorAll(".done").forEach(e => e.hidden = !toggleDone)
  }

</script>
<style>

.submit input:not([type="submit" i]){
  display: block;
  min-width: 300px;
  margin-bottom: 1em;
  margin-top: 0.5em;
}

.submit textarea {
  width: 100%;
  height: 250px;
}

.grid-note{
    grid-template-columns: 1fr;
   word-wrap: break-word;
  font-size: 14px;

}

.content {
  max-width: 450px;
}

.done .content {
  color: lightgrey;
  text-decoration: line-through;
}

.done .content a {
  text-decoration: line-through;
}

.note > hr {
    height: 1px;
    background-color: whitesmoke;
    border: none;  

}

.prev-next {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1em;
}

.prev-next > a {
  text-decoration: none;
}

.controls {
  display: flex;
  justify-content: end;
}

.controls > * {
  margin-left: 1em;
}

@media screen and (max-width: 400px) {
  .grid-note{
      grid-template-columns: 1fr;
  }

  .note {
    max-width: 300px;
  }

.controls {
  flex-direction: row;
}
}
  .emoji-button {
    cursor: pointer;
  }

.done-toggle {
  border-radius: 25px;
  width: 50px;
  height: 25px;
  position: relative;
  background: green;
  cursor: pointer;
  margin-left: auto;
  margin-bottom: 2em;
}

.slider {
  position: absolute;
  left: -0.5px;
  top: -0.5px;
  width: 25px;
  height: 25.6px;
  border-radius: 25px;
  background-color: white;
  transition: .2s;
}

.toggled {
  left:unset;
  transform: translateX(25.5px)
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
    <div class="done-toggle" on:click={()=>toggleDoneFilter()}> 
      <div class="slider" class:toggled={!toggleDone}/>
      </div>
    {#each notes as note}
      <div class="note" class:done={note.done}>
        <a name={note.id} />
        <div class="controls">
          <input
            name="done"
            type="checkbox"
            checked={note.done}
            on:change={() => toggle(note.id)}
          />
          <div class="emoji-button" on:click={()=>toggleEdit(note.id)}>
            {#if editingId === note.id}
              &#10060;
            {:else}
              &#128397;
            {/if}
          </div>
          <div class="emoji-button" on:click={() => location.href=`/note/${note.id}/delete`}>&#x1F5D1;</div>
        </div>
          <div class="content">
        {#if editingId === note.id}
        <div class="note submit" >
          <textarea type="text" name="body" required bind:value={editingBody}/>
          <input type="text" name="tags" placeholder="use comma 'seperated values'" bind:value={editingTags} autocorrect="off" autocapitalize="none"/>
          <input type="submit" value="Submit" on:click={()=>handleEdit(note.id)}/>
        </div>
        {:else}
        {@html marked(note.body)}
          <Metadata tags={note.tags}/>
        {/if}
         </div>
        <hr>
      </div>
    {/each}
  </div>
  <div class="prev-next">
    <a href={`/t/${previousDay}`}>&larr;</a>
    <a href={`/t/${nextDay}`}>&rarr;</a>
  </div>
</main>
