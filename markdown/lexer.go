package markdown

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	name  string    // used for error reports
	input string    // string being scanned
	start int       // start position of this item
	pos   int       // current position in the input
	width int       // width of the last rune read from input
	items chan item // channel of the scanned items
}

const eof = rune(0)
const ws = ' '

type stateFn func(*lexer) stateFn

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemHeader
	itemText
	itemNewLine
	itemList
	itemBold
)

const header = "#"
const eol = "\n"
const list = "-"
const bold = "*"

type item struct {
	typ itemType
	val string
}

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}

	go l.run()

	return l, l.items
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}

	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func lexNewLine(l *lexer) stateFn {
	if l.accept(eol) {
		l.emit(itemNewLine)
	}
	return lexText
}

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], list) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexList
		}
		if strings.HasPrefix(l.input[l.pos:], eol) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexNewLine
		}
		if strings.HasPrefix(l.input[l.pos:], header) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexHeader
		}
		if strings.HasPrefix(l.input[l.pos:], bold) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBold
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func lexHeader(l *lexer) stateFn {
	if l.accept(header) {
		l.acceptRun(header)
	}
	if l.peek() == ws {
		l.emit(itemHeader)
		l.next()
		l.ignore()
	}
	return lexText
}

func lexBold(l *lexer) stateFn {
	if l.accept(bold) {
		l.emit(itemBold)
	}
	return lexText
}

func lexList(l *lexer) stateFn {
	if l.accept(list) && l.peek() == ws {
		l.emit(itemList)
		l.next()
		l.ignore()
	}
	return lexText
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) skip() {
	l.pos += 1
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}

	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}

	return fmt.Sprintf("%q", i.val)
}
