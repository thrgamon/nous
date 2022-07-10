<script>
  import TodoList from "./TodoList.svelte"
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
  {#if todos !== null} 
   <TodoList notes={todos}/>
  {:else}
    You have done all your todos!
  {/if} 
{:catch error}
  There was a problem loading todos
  {error.message}
{/await}
