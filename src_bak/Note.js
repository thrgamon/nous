import './Note.css';
import React, { memo } from "react";
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkBreaks from 'remark-breaks'

function Note({ note }) {
  return (
    <div className="note prose-2xl" key={note.id}>
      <ReactMarkdown children={note.body} remarkPlugins={[remarkGfm, remarkBreaks]} />
      <div className="tags">
        {note.tags.map(tag => <span>{tag}</span>)}
      </div>
    </div>
  )
}

export default Note;
export const RenderedNote = memo(Note);