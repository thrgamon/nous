import type { NextPage } from 'next'
import React, { useState, useCallback } from "react";
import useSWR, { useSWRConfig } from 'swr'
import useSWRMutation from 'swr/mutation'

import Note from "../components/Note"
import ConfiguredCodeMirror from "../components/ConfiguredCodeMirror"

const Home: NextPage = () => {
  const { mutate } = useSWRConfig()
  const fetcher = (input: RequestInfo | URL, init?: RequestInit) => fetch(input, init).then(res => res.json())
  async function sendRequest(url, { arg }) {
    return fetch(url, {
      method: 'POST',
      body: JSON.stringify(arg)
    })
  }
  const [value, setValue] = useState("");
  const onChange = useCallback((value, viewUpdate) => {
    setValue(value)
  }, []);
  const resetNote = () => setValue("")
  const { trigger: createNote } = useSWRMutation('/api/note', sendRequest, { onSuccess: resetNote })
  const { data: notes, error } = useSWR('/api/notes', fetcher)

  if (error) return <div>failed to load</div>
  if (!notes) return <div>loading...</div>

  const handleKeyDown = (e) => {
    if (e.metaKey && (e.keyCode == 10 || e.keyCode == 13)) {
      mutate('/api/notes', createNote({ body: value, tags: ['hello'] }), {
        populateCache: (res, notes) => {
          res.json().then((note) => {return [note, ...notes] })
        },
        revalidate: false
      })
    }
  }

  return (
    <div className="grid grid-cols-3">
      <div className="col-start-2" data-color-mode="light" onKeyDown={handleKeyDown}>
        <ConfiguredCodeMirror value={value} onChange={onChange} />
      </div>
      <div className="col-start-2">
        {console.log(notes)}
        {notes.map((note: any) => (
          <div className="note" key={note.id}>
            <Note note={note} />
          </div>
        ))}
      </div>
    </div>
  )
}

export default Home
