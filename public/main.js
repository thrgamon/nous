const checkboxes = document.querySelectorAll("input[name=done]");

checkboxes.forEach(checkbox => {
  checkbox.addEventListener('change', (event) => {
    const id = checkbox.value
    postData('/api/done', { Id: id })
    .then(response => {
      const note = document.querySelector(`[data-id='${id}']`);
      note.classList.toggle('done')
    })
    .catch(error => {
      checkbox.checked = !checkbox.checked
    })

    event.preventDefault()
  });
});

const noteContent = document.querySelectorAll("[data-note-content]")

for (i = 0; i < noteContent.length; i++) {
  const note = noteContent[i]
  const links = note.querySelectorAll('a')

  if (links.length == 0) {
    continue
  } else {
    links.forEach(link => {
      const details = document.createElement("details")
      const iframe = document.createElement("iframe")
      iframe.src = link.href
      details.append(iframe)
      note.append(details)
    })
  }
}

// Example POST method implementation:
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

