<script>
  import Home from "./Home.svelte"
  import Notes from "./Notes.svelte"
  import Todos from "./Todos.svelte"
  import Reading from "./Reading.svelte"
  import { Router, Route, Link } from "svelte-navigator";
  import dayjs from 'dayjs'

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
    <h1><a href="/">Nous</a></h1>
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
      <Home currentDay={dayjs().format('YYYY-MM-DD')}/>
    </Route>
    <Route path="search">
      <Notes notes={notes}/>
    </Route>
    <Route path="t/:date" let:params>
      <Home currentDay={params.date}/>
    </Route>
    <Route path="todos">
      <Todos/>
    </Route>
    <Route path="readings">
      <Reading/>
    </Route>
  </main>
</Router>
