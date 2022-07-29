package notes

import (
	"errors"
	"regexp"
)

func ExtractPeople(text string) (people []string) {
	re := regexp.MustCompile(`\B@([\w]+)`)

	for _, match := range re.FindAllString(text, -1) {
		people = append(people, match[1:])
	}

	return people
}

func ToggleTodo(body string, index int) (string, error) {
	re := regexp.MustCompile(`- \[[ xX]\]`)
	tasks := re.FindAllStringIndex(body, index+1)
	if len(tasks) == 0 {
		return "", errors.New("No todos found")
	}
	bodyIndexes := tasks[index]
	pre := body[:bodyIndexes[0]]
	todo := body[bodyIndexes[0]:bodyIndexes[1]]
	post := body[bodyIndexes[1]:]

	if todo == "- [x]" || todo == "- [X]" {
		todo = "- [ ]"
	} else {
		todo = "- [x]"
	}

	return pre + todo + post, nil
}
