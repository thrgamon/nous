<script>
  import Notes from "./Notes.svelte"
  let readings = getReadings();

  async function getReadings() {
    const response = await fetch('/api/readings');
    return await response.json();
  }

</script>
  <style>

</style>


{#await readings}
  Loding readings
{:then readings}
  <Notes notes={readings}/>
{:catch error}
  There was a problem loading readings
  {error.message}
{/await}
