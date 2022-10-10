import React from "react"
import CodeMirror from '@uiw/react-codemirror';
import { EditorView, keymap, highlightSpecialChars } from "@codemirror/view"
import { defaultKeymap, emacsStyleKeymap, indentWithTab, history } from "@codemirror/commands"
import { bracketMatching, indentOnInput, syntaxHighlighting, defaultHighlightStyle } from "@codemirror/language"
import { closeBrackets } from "@codemirror/autocomplete"
import { markdown } from "@codemirror/lang-markdown"

const ConfiguredCodeMirror = ({value, onChange}) => {
  return <CodeMirror
    height="300px"
    value={value}
    autoFocus={true}
    basicSetup={false}
    extensions={[syntaxHighlighting(defaultHighlightStyle, { fallback: true }), EditorView.lineWrapping, bracketMatching(), history(), indentOnInput(), closeBrackets(), highlightSpecialChars(), keymap.of([defaultKeymap, emacsStyleKeymap, indentWithTab]), markdown({ markdownLanguage: 'GFM' })]}
    onChange={onChange}
  />
};

export default ConfiguredCodeMirror;
