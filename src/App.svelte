<script>
  import Notes from "./Notes.svelte"
  import Todos from "./Todos.svelte"
  import Reading from "./Reading.svelte"
  import { Router, Route, Link } from "svelte-navigator";

  export let previousDay;
  export let nextDay;
  export let currentDay;
  export let notes;

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
<Router>
  <header>
    <h1><Link to="/">Nous</Link></h1>
    <nav>
      <Link to="todos">Todos</Link>
      <Link to="readings">Readings</Link>
      <form action="/search" method="get">
        <input type="text" name="query" required placeholder="search">
        <input type="submit" value="Search">
      </form>
    </nav>
  </header>
  <main>
    <Route path="/">
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
      <Notes notes={notes}/>
      <div class="prev-next">
        <a href={`/t/${previousDay}`}>&larr;</a>
        <a href={`/t/${nextDay}`}>&rarr;</a>
      </div>
    </Route>
    <Route path="todos">
      <Todos/>
    </Route>
    <Route path="readings">
      <Reading/>
    </Route>
  </main>
</Router>
