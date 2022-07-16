package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// In production we want to read the cached templates, whereas in development
	// we want to interpret them every time to make it easier to change
	if ENV == Production {
		err := Templates[tmpl].Execute(w, data)

		if err != nil {
			Logger.Println(err.Error())
			handleUnexpectedError(w, err)
			return
		}
	} else {
		template := template.Must(template.ParseFiles("views/"+tmpl+".html", "views/_header.html", "views/_footer.html"))
		err := template.Execute(w, data)

		if err != nil {
			Logger.Println(err.Error())
			handleUnexpectedError(w, err)
			return
		}
	}
}

func cacheTemplates() map[string]*template.Template {
	re := regexp.MustCompile(`[a-zA-Z\/]*\.html`)
	templates := make(map[string]*template.Template)
	// Walk the template directory and parse all templates that aren't fragments
	err := filepath.WalkDir("views",
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path != "_header.html" && path != "_footer.html" && re.MatchString(path) {
				normalisedPath := strings.TrimSuffix(strings.TrimPrefix(path, "views/"), ".html")
				templates[normalisedPath] = template.Must(
					template.ParseFiles(path, "views/_header.html", "views/_footer.html"),
				)
			}

			return nil
		})

	if err != nil {
		log.Fatal(err.Error())
	}

	return templates
}
