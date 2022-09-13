import './App.css';
import { memo } from "react";
import MDEditor from '@uiw/react-md-editor';

function Note({body}) {
  return ( <MDEditor.Markdown source={body}/>)
}

export default Note;
export const RenderedNote = memo(Note);
