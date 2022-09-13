import './App.css';
import { useState, useEffect, memo } from "react";
import { RenderedNote } from "./Note.js"
import MDEditor from '@uiw/react-md-editor';

function App() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [value, setValue] = useState("");

  useEffect(() => {
    fetch(`/api/notes`)
      .then((response) => {
        if (!response.ok) {
          throw new Error(
            `This is an HTTP error: The status is ${response.status}`
          );
        }
        return response.json();
      })
      .then((actualData) => setData(actualData))
      .catch((err) => {
        alert(err)
      })
      .finally(setLoading(false))
  }, []);

  const handleKeyDown = (e) => {
    if (e.metaKey && (e.keyCode == 10 || e.keyCode == 13)) {
      createNewNote({body: e.target.value, tags: ['hello']})
    }
  }

  const createNewNote = (note) => {
    fetch(`/api/note`, { method: 'post', body: JSON.stringify(note) })
      .then((response) => {
        if (!response.ok) {
          throw new Error(
            `This is an HTTP error: The status is ${response.status}`
          );
        }
      })
      .then(() => {
        setData([note, ...data])
        setValue("")
      })
      .catch((err) => {
        alert(err)
      })
  }

  return (
    <div>
      <MDEditor
        onKeyDown={handleKeyDown}
        value={value}
        onChange={setValue}
        autoFocus
      />
      <div>
        {loading && <div>A moment please...</div>}
        {error && (
          <div>{`There is a problem fetching the post data - ${error}`}</div>
        )}

      </div>
      <ul>
        {data &&
          data.map((note) => (
            <div key={note.id}>
              <RenderedNote body={note.body} />
              <hr />
            </div>
          ))}
      </ul>
    </div>
  );
}

export default App;
