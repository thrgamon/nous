<script>
  import Notes from "./Notes.svelte"
  import dayjs from 'dayjs'

  export let currentDay;
  let nextDay = dayjs(currentDay, 'YYYY-MM-DD').add(1, 'day').format('YYYY-MM-DD')
  let previousDay = dayjs(currentDay, 'YYYY-MM-DD').subtract(1, 'day').format('YYYY-MM-DD')
  let notes = getNotes();

  async function getNotes() {
    const response = await fetch(`/api/notes?from=${currentDay}&to=${nextDay}`);
    return await response.json();
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

.prev-next {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1em;
}

.prev-next > a {
  text-decoration: none;
}


</style>
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

{#await notes}
  Loding notes
{:then notes}
  {#if notes === null}
    No Notes Yet
  {:else}
    <Notes notes={notes}/>
  {/if}
{:catch error}
  {console.log(notes)}
  There was a problem loading notes
  {error.message}
{/await}
<div class="prev-next">
  <a href={`/t/${previousDay}`}>&larr;</a>
  <a href={`/t/${nextDay}`}>&rarr;</a>
</div>
