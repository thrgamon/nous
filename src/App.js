import './App.css';
import { useState, useEffect, useCallback } from "react";
import { RenderedNote } from "./Note.js"
import Scali  from "./Scali.js"
import ConfiguredCodeMirror  from "./ConfiguredCodeMirror.js"

function App() {
  const [mode, setMode] = useState(0)
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [value, setValue] = useState("");

  useEffect(() => {
    const params = new URLSearchParams({
    from: "2022-07-01",
    to: "2022-08-01",
});
    fetch('/api/notes'+"?"+params)
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

const onChange = useCallback((value, viewUpdate) => {
    setValue(value)
  }, []);

  const handleKeyDown = (e) => {
    if (e.metaKey && (e.keyCode == 10 || e.keyCode == 13)) {
      createNewNote({ body: value, tags: ['hello'] })
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

  const notesView = () => {
  return (
    <div className="notes" data-color-mode="light">
      <div onKeyDown={handleKeyDown}>
      <ConfiguredCodeMirror value={value} onChange={onChange} />
    </div>
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

  const scaliView = () => {
    return <Scali />
  }

  const renderMode = () => {
    if (mode===0) {
      return notesView()
    } else {
      return scaliView()
    }
  }

  return (
    <div>
      <button
    onClick={() => setMode(mode === 0 ? 1 : 0)}
    >SwitchMode</button>
    {renderMode()}
    </div>
  );
}

export default App;
