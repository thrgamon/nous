<script>
  import { marked } from "marked";

  export let previousDay;
  export let nextDay;
  export let notes;

  async function postData(url = '', data = {}, errorCallback) {
    // Default options are marked with *
    const response = await fetch(url, {
      method: 'POST',
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
      <div class="flex-center">
        <input
          name="done"
          type="checkbox"
          checked={note.Done}
          on:change={() => toggle(note.ID)}
        />
      </div>
      <div class="note" class:done={note.Done}>
        <a name={note.ID} />
        {@html marked(note.Body)}
        <hr />
      </div>
      <div class="metadata">
        <ul class="tags text-subdued">
          {#each note.Tags as tag}
            <li class="tag">{tag}</li>
          {/each}
        </ul>
      </div>
      <div class="flex-center">
        <a href="/note/{note.ID}/delete">delete</a>
      </div>
    {/each}
  </div>
</main>
