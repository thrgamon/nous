import './App.css';
import { memo } from "react";
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkBreaks from 'remark-breaks'

function Note({body}) {
  return ( <ReactMarkdown children={body} remarkPlugins={[remarkGfm, remarkBreaks]} />)
}

export default Note;
export const RenderedNote = memo(Note);
