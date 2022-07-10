<script>
  import { marked } from "marked";

  marked.setOptions({
    gfm: true,
    breaks: true,
    smartLists: true,
});
  export let notes;

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

</script>
  <style>

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

.todo > * {
  display: inline-block;
}

.note > hr {
    height: 1px;
    background-color: whitesmoke;
    border: none;  

}

@media screen and (max-width: 400px) {
  .grid-note{
      grid-template-columns: 1fr;
  }

  .note {
    max-width: 300px;
  }
}
</style>

  <div class="grid-note">
    {#each notes as note}
      <div class="note todo" class:done={note.done}>
          <input
            name="done"
            type="checkbox"
            checked={note.done}
            on:change={() => toggle(note.id)}
          />
          <div class="content">
        {@html marked(note.body)}
         </div>
        <hr>
      </div>
    {/each}
  </div>
