package main

import (
	"embed"
	"html/template"
	"io/fs"
  "fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
  "path/filepath"

	"github.com/thrgamon/learning_rank/env"

	"context"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

var REPO *ResourceRepo
var DB *pgxpool.Pool

var Templates map[string]*template.Template

//go:embed views/*
var views embed.FS

func main() {
  DB = initDB()
  defer DB.Close()

  REPO = NewResourceRepo(DB)
  cacheTemplates()

  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler)
  r.HandleFunc("/resource", AddResourceHandler)
  r.HandleFunc("/up/{resourceId:[0-9]+}", UpvoteHandler)
  r.HandleFunc("/down/{resourceId:[0-9]+}", DownvoteHandler)
  r.PathPrefix("/public/").HandlerFunc(serveResource)

  srv := &http.Server{
    Handler: handlers.CombinedLoggingHandler(os.Stdout, r),
    Addr:    "0.0.0.0:" + env.GetEnvWithFallback("PORT", "8080"),
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }
  println("Server listening")
	log.Fatal(srv.ListenAndServe())
}

func cacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "max-age=60000, public")
		next.ServeHTTP(w, r)
	})
}

func serveResource(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(filepath.Join(".", r.URL.Path))
	if err != nil {
		http.Error(w, r.RequestURI, http.StatusNotFound)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, r.RequestURI, http.StatusNotFound)
		return
	}
	modTime := fi.ModTime()

	http.ServeContent(w, r, r.URL.Path, modTime, f)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  resources, err := REPO.GetAll(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

  RenderTemplate(w, "home", resources)
}

func UpvoteHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  resourceId, _ := strconv.ParseUint(vars["resourceId"], 10, 64)
  err := REPO.Upvote(r.Context(), uint(resourceId))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

  http.Redirect(w, r, "/", 303)
}

func AddResourceHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()

	name := r.FormValue("name")
	link := r.FormValue("link")

  err := REPO.Add(r.Context(), link, name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

  http.Redirect(w, r, "/", 303)
}

func DownvoteHandler(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  resourceId, _ := strconv.ParseUint(vars["resourceId"], 10, 64)
  err := REPO.Downvote(r.Context(), uint(resourceId))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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

func initDB() *pgxpool.Pool{
  conn, err := pgxpool.Connect(context.TODO(), os.Getenv("DATABASE_URL"))

  if err != nil {
    log.Fatal(err)
  }

  return conn
}
