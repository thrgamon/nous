<script>
  export let tags
</script>
  <style>
.tags {
  padding-inline-start: 0;
}

.text-subdued {
    font-size: 0.8em;
  color: darkgray;
}

.tag {
    display: inline-block;
}
.tag ~ .tag::before{
  margin-left: -0.25em;
  content: ", ";
}

.metadata {
    display: flex;
    align-items: right;
    justify-content: top;
}


</style>

<div class="metadata">
<ul class="tags text-subdued">
  {#each tags as tag}
    <li class="tag">
      <a href={`/search?query=${tag}`}>{tag}</a>
    </li>
  {/each}
</ul>
    </div>
