package markdown

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeader1(t *testing.T) {
	buf := bytes.Buffer{}
	text := "# The Title\n"
	want := "<h1>The Title</h1>"
	_, itemChan := lex("markdown", text)
  parse(itemChan, &buf)
	got := buf.String()

	assert.Equal(t, want, got)
}

func TestParseHeader2(t *testing.T) {
	buf := bytes.Buffer{}
	text := "## The Title\n"
	want := "<h2>The Title</h2>"
	_, itemChan := lex("markdown", text)
  parse(itemChan, &buf)
	got := buf.String()

	assert.Equal(t, want, got)
}

func TestParseHeader3(t *testing.T) {
	buf := bytes.Buffer{}
	text := "##The Title\n"
	want := "##The Title\n"
	_, itemChan := lex("markdown", text)
  parse(itemChan, &buf)
	got := buf.String()

	assert.Equal(t, want, got)
}

func TestParseHeader4(t *testing.T) {
	buf := bytes.Buffer{}
	text := "## The Title"
	want := "<h2>The Title</h2>"
	_, itemChan := lex("markdown", text)
  parse(itemChan, &buf)
	got := buf.String()

	assert.Equal(t, want, got)
}


func TestParseList(t *testing.T) {
	buf := bytes.Buffer{}
	text := "- list 1\n- list 2\n\n"
  want := "<ul>\n<li>list 1</li>\n<li>list 2</li>\n</ul>\n"
	_, itemChan := lex("markdown", text)
  parse(itemChan, &buf)
	got := buf.String()

	assert.Equal(t, want, got)
}

func TestParseList1(t *testing.T) {
	buf := bytes.Buffer{}
	text := "- list 1\n- list 2"
  want := "<ul>\n<li>list 1</li>\n<li>list 2</li>\n</ul>\n"
	_, itemChan := lex("markdown", text)
  parse(itemChan, &buf)
	got := buf.String()

	assert.Equal(t, want, got)
}

