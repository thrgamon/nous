<script>
  import Notes from "./Notes.svelte"
  let todos = getTodos();

  async function getTodos() {
    const response = await fetch('/api/todos');
    return await response.json();
  }

</script>
  <style>

</style>


{#await todos}
  Loding todos
{:then todos}
  <Notes notes={todos}/>
{:catch error}
  There was a problem loading todos
  {error.message}
{/await}
