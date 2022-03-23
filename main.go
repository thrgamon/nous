package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var REPO *ResourceRepo

var Templates map[string]*template.Template

//go:embed public/*
var public embed.FS

//go:embed views/*
var views embed.FS

func main() {
  REPO = NewResourceRepo()
  cacheTemplates()

  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler)
  r.HandleFunc("/up/{resourceId:[0-9]+}", UpvoteHandler)
  r.HandleFunc("/down/{resourceId:[0-9]+}", DownvoteHandler)
  r.PathPrefix("/public/").Handler(http.FileServer(http.FS(public)))

  srv := &http.Server{
    Handler: handlers.CombinedLoggingHandler(os.Stdout, r),
    Addr:    "0.0.0.0:8080",
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }

	log.Fatal(srv.ListenAndServe())
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  RenderTemplate(w, "home", REPO.storage)
}

func UpvoteHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  resourceId, _ := strconv.ParseUint(vars["resourceId"], 10, 64)
  REPO.Upvote(uint(resourceId))
  http.Redirect(w, r, "/", 303)
}

func DownvoteHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  resourceId, _ := strconv.ParseUint(vars["resourceId"], 10, 64)
  REPO.Downvote(uint(resourceId))
  http.Redirect(w, r, "/", 303)
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := Templates[tmpl].Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func cacheTemplates() {
  re := regexp.MustCompile(`^[a-zA-Z\/]*\.html`)
	templates := make(map[string]*template.Template)
	// Walk the template directory and parse all templates that aren't fragments
	err := fs.WalkDir(views, ".",
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if re.MatchString(path) {
				normalisedPath := strings.TrimSuffix(strings.TrimPrefix(path, "views/"), ".html")
				templates[normalisedPath] = template.Must(
					template.ParseFS(
						views,
						path,
					),
				)
			}

			return nil
		})

	if err != nil {
		log.Fatal(err.Error())
	}

	// Assign to global variable so we can access it when rendering templates
	Templates = templates

}
