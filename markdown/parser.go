package markdown

import (
	"bytes"
	"fmt"
)

type parser struct {
	items  chan item // channel of the scanned items
	output *bytes.Buffer
}

type contextType int

const (
  inText contextType = iota
  inList
)

type Tag interface {
  Open() string
  Close() string
  Tag() string
}

type HeaderTag struct {
  level int
}

func (h *HeaderTag) Open() string {
 return fmt.Sprintf("<h%d>", h.level)
}

func (h *HeaderTag) Close() string{
 return fmt.Sprintf("</h%d>", h.level)
}

func (h *HeaderTag) Tag() string{
 return "h"
}

type UnorderedListTag struct{}

func (ul *UnorderedListTag) Open() string {
 return fmt.Sprintf("<ul>")
}

func (ul *UnorderedListTag) Close() string{
 return fmt.Sprintf("</ul>")
}

func (ul *UnorderedListTag) Tag() string{
 return "ul"
}

type ListItemTag struct{}

func (li *ListItemTag) Open() string {
 return fmt.Sprintf("<li>")
}

func (li *ListItemTag) Close() string{
 return fmt.Sprintf("</li>")
}

func (li *ListItemTag) Tag() string{
 return "li"
}

func parse(items chan item, buf *bytes.Buffer) *bytes.Buffer {
	p := &parser{
		items:  items,
		output: buf,
	}

  context := inText
  stack := NewStack()

  for item := range p.items {
    switch item.typ {
  case itemText: 
		fmt.Fprintf(p.output, item.val)
  case itemHeader: 
      h := &HeaderTag{len(item.val)}
      stack.Push(h)
      fmt.Fprintf(p.output, h.Open())
  case itemList: 
      if context != inList {
        ul := &UnorderedListTag{}
        stack.Push(ul)
        context = inList
        fmt.Fprintf(p.output, ul.Open())
		    fmt.Fprintf(p.output, "\n")
      }
      li := &ListItemTag{}
      stack.Push(li)
      fmt.Fprintf(p.output, li.Open())
  case itemNewLine: 
      if stack.Empty() {
		    fmt.Fprintf(p.output, "\n")
      } else {
      switch stack.Peek().Tag() {
      case "h1":
		    fmt.Fprintf(p.output, stack.Pop().Close())
      case "li":
		    fmt.Fprintf(p.output, stack.Pop().Close())
		    fmt.Fprintf(p.output, "\n")
      case "ul":
		    fmt.Fprintf(p.output, stack.Pop().Close())
		    fmt.Fprintf(p.output, "\n")
        context = inText
      }
    }
  case itemEOF:
      if stack.Empty() {
        break
      }
      switch stack.Peek().Tag() {
      case "h":
		    fmt.Fprintf(p.output, stack.Pop().Close())
      case "li":
		    fmt.Fprintf(p.output, stack.Pop().Close())
		    fmt.Fprintf(p.output, "\n")
		    fmt.Fprintf(p.output, stack.Pop().Close())
		    fmt.Fprintf(p.output, "\n")
      }
  }
  }

  return buf
}

type Stack struct {
  s []Tag
}

func NewStack() *Stack {
  return &Stack{[]Tag{}}
}

func (s *Stack) Push(item Tag) {
  s.s = append(s.s, item)
}

func (s *Stack) Empty() bool{
  n := len(s.s)
  return n == 0
}

func (s *Stack) Peek() Tag {
  n := len(s.s) - 1
  return s.s[n]
}

func (s *Stack) Pop() (item Tag){
  n := len(s.s) -1
  item = s.s[n]
  s.s = s.s[:n]

  return item
}
