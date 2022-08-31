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

document.querySelector("#editor").addEventListener("keydown", (event) => {
if ((event.keyCode == 10 || event.keyCode == 13) && event.metaKey) {
  const form = event.target.closest('form')
  const formTrigger = form.querySelector("button.submit");
  const submitEvent = new SubmitEvent("submit", { submitter: formTrigger });
  form.dispatchEvent(submitEvent);
}
})


