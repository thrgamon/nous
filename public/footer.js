function hydrateCheckboxes() {
  document.querySelectorAll('.content input').forEach(input => input.closest('.content').querySelectorAll('input').forEach((x, index) => {
  x.disabled = false;
  id = x.closest('.note').dataset.noteId;
  x.setAttribute("hx-put", `/note/${id}/todo/${index}`);
  x.setAttribute("hx-target", "closest .note")
  htmx.process(x)
}))
}

hydrateCheckboxes()

document.addEventListener("htmx:afterSettle", hydrateCheckboxes)
