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
  inHeader
)

type Tag interface {
  Open() string
  Close() string
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

func parse(items chan item, buf *bytes.Buffer) *bytes.Buffer {
	p := &parser{
		items:  items,
		output: buf,
	}

  stack := NewStack()

  for item := range p.items {
    switch item.typ {
  case itemText: 
		fmt.Fprintf(p.output, item.val)
  case itemHeader: 
      h := &HeaderTag{len(item.val)}
      stack.Push(h)
      fmt.Fprintf(p.output, h.Open())
  case itemNewLine: 
      if stack.Empty() {
		    fmt.Fprintf(p.output, "\n")
      } else {
      switch stack.Peek().(type) {
      case *HeaderTag:
		    fmt.Fprintf(p.output, stack.Pop().Close())
      }
    }
  case itemEOF:
      if stack.Empty() {
        break
      }
      switch stack.Peek().(type) {
      case *HeaderTag:
		    fmt.Fprintf(p.output, stack.Pop().Close())
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
  n := len(s.s) -1
  return s.s[n]
}

func (s *Stack) Pop() (item Tag){
  n := len(s.s) -1
  item = s.s[n]
  s.s = s.s[:n]

  return item
}
