package main

import (
  "github.com/gorilla/mux"
  "net/http"
  "html/template"
  "log"
  "time"
  "strconv"
  "embed"
  )

var REPO *ResourceRepo

//go:embed public/*
var public embed.FS

func main() {
  REPO = NewResourceRepo()

  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler)
  r.HandleFunc("/up/{resourceId:[0-9]+}", UpvoteHandler)
  r.HandleFunc("/down/{resourceId:[0-9]+}", DownvoteHandler)
  r.PathPrefix("/public/").Handler(http.FileServer(http.FS(public)))

  srv := &http.Server{
    Handler: r,
    Addr:    "127.0.0.1:8080",
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
  }

	log.Fatal(srv.ListenAndServe())
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
  RenderTemplate(w, "home.html", REPO.storage)
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
  t := template.Must(template.ParseFiles(tmpl))
	err := t.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
