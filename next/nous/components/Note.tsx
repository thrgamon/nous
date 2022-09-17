import React, { useMemo } from "react";
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkBreaks from 'remark-breaks'

function Note({ note }) {
const renderMarkdown = (note) => <ReactMarkdown children={note.body} remarkPlugins={[remarkGfm, remarkBreaks]} />
const markdown = useMemo(() => renderMarkdown(note), [note]);

  return (
    <div className="note prose" key={note.id}>
    <div>
      {markdown}
    </div>
      <div className="tags">
        {note.tags.map(tag => <span>{tag}</span>)}
      </div>
    </div>
  )
}

export default Note;

