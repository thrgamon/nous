import { useState, useEffect, useCallback } from "react";
import {EditorView, keymap, highlightSpecialChars} from "@codemirror/view"
import {defaultKeymap, emacsStyleKeymap, indentWithTab, history} from "@codemirror/commands"
import {bracketMatching, indentOnInput, syntaxHighlighting, defaultHighlightStyle} from "@codemirror/language"
import {closeBrackets} from "@codemirror/autocomplete"
import {markdown} from "@codemirror/lang-markdown"

export default function useCodeMirror(onChange) {
  const [element, setElement] = useState();

  const ref = useCallback((node) => {
    if (!node) return;

    setElement(node);
  }, []);

  useEffect(() => {
    if (!element) return;

    const view = new EditorView({
      lineWrapping: true,
      extensions: [syntaxHighlighting(defaultHighlightStyle, {fallback: true}), EditorView.lineWrapping, bracketMatching(), history(), indentOnInput(), closeBrackets(), highlightSpecialChars(), keymap.of([defaultKeymap, emacsStyleKeymap, indentWithTab]), markdown({markdownLanguage: 'GFM'})],
      parent: element,
      onChange
    });

    return () => view?.destroy();
  }, [element]);

  return { ref };
}
