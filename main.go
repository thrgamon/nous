package main

import (
	"context"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/thrgamon/learning_rank/env"
	"github.com/thrgamon/nous/repo"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Environment int

const (
	Production Environment = iota + 1
	Development
)

var DB *pgxpool.Pool
var Templates map[string]*template.Template
var Log *log.Logger
var ENV Environment

func main() {
	if env.GetEnvWithFallback("ENV", "production") == "development" {
		ENV = Development
	} else {
		ENV = Production
	}

	DB = initDB()
	defer DB.Close()

	cacheTemplates()

	Log = log.New(os.Stdout, "logger: ", log.Lshortfile)

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
  r.HandleFunc("/t/{date}", HomeSinceHandler)
	r.HandleFunc("/submit", SubmitHandler)
	r.HandleFunc("/search", SearchHandler)
	r.PathPrefix("/public/").HandlerFunc(serveResources)
	r.HandleFunc("/note", AddNoteHandler)
	r.HandleFunc("/note/{id:[0-9]+}/delete", DeleteNoteHandler)
	r.HandleFunc("/note/toggle", ToggleNoteHandler)

	srv := &http.Server{
		Handler:      handlers.CombinedLoggingHandler(os.Stdout, r),
		Addr:         "0.0.0.0:" + env.GetEnvWithFallback("PORT", "8080"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	Log.Println("Server listening")
	log.Fatal(srv.ListenAndServe())
}

type PageData struct {
	Notes []repo.Note
}

func HomeSinceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

  parsedTime, err := time.Parse(time.RFC3339, date+"T00:00:00+11:00")

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	noteRepo := repo.NewNoteRepo(DB)
	notes, err := noteRepo.GetAllSince(r.Context(), parsedTime)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	RenderTemplate(w, "home", pageData)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(DB)
	notes, err := noteRepo.GetAllSince(r.Context(), time.Now())

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	RenderTemplate(w, "home", pageData)
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, "submit", PageData{})
}

func ViewNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteId := vars["noteId"]

	noteRepo := repo.NewNoteRepo(DB)
	note, err := noteRepo.Get(r.Context(), repo.NoteID(noteId))

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: []repo.Note{note}}
	RenderTemplate(w, "view", pageData)
}

func ToggleNoteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id := r.FormValue("id")

	noteRepo := repo.NewNoteRepo(DB)
	err := noteRepo.ToggleDone(r.Context(), repo.NoteID(id))

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	http.Redirect(w, r, "/#"+id, http.StatusSeeOther)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	noteRepo := repo.NewNoteRepo(DB)
	err := noteRepo.Delete(r.Context(), repo.NoteID(id))

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func AddNoteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	body := r.FormValue("body")
	tags := r.FormValue("tags")

	noteRepo := repo.NewNoteRepo(DB)
	err := noteRepo.Add(r.Context(), body, tags)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	query := r.FormValue("query")

	noteRepo := repo.NewNoteRepo(DB)
	notes, err := noteRepo.Search(r.Context(), query)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	RenderTemplate(w, "home", pageData)
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// In production we want to read the cached templates, whereas in development
	// we want to interpret them every time to make it easier to change
	if ENV == Production {
		err := Templates[tmpl].Execute(w, data)

		if err != nil {
			handleUnexpectedError(w, err)
			return
		}
	} else {
		template := template.Must(template.ParseFiles("views/"+tmpl+".html", "views/_header.html", "views/_footer.html"))
		err := template.Execute(w, data)

		if err != nil {
			handleUnexpectedError(w, err)
			return
		}
	}
}

func cacheTemplates() {
	re := regexp.MustCompile(`^[a-zA-Z\/]*\.html`)
	templates := make(map[string]*template.Template)
	// Walk the template directory and parse all templates that aren't fragments
	err := filepath.WalkDir("views",
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if re.MatchString(path) {
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

	// Assign to global variable so we can access it when rendering templates
	Templates = templates

}

func initDB() *pgxpool.Pool {
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	return conn
}

// Handler for serving static assets with modified time to help
// caching
func serveResources(w http.ResponseWriter, r *http.Request) {
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

func handleUnexpectedError(w http.ResponseWriter, err error) {
	http.Error(w, "There was an unexpected error", http.StatusInternalServerError)
	Log.Println(err.Error())
}
