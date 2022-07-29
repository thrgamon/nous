package notes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPeople(t *testing.T) {
	text := "this is some text with @hannah and \n @tom in it as well as an email mike@tomgamon.com"
	want := []string{"hannah", "tom"}

	got := ExtractPeople(text)
	assert.Equal(t, want, got)
}

func TestToggleTodo(t *testing.T) {
	text := "this is some text with @hannah and \n - [ ] todo"
	want := "this is some text with @hannah and \n - [x] todo"

	got, err := ToggleTodo(text, 0)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
