package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/thrgamon/go-utils/env"
	urepo "github.com/thrgamon/go-utils/repo/user"
	"github.com/thrgamon/go-utils/web/authentication"
	"github.com/thrgamon/nous/repo"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Environment int

const (
	Production Environment = iota + 1
	Development
)

var DB *pgxpool.Pool
var Templates map[string]*template.Template
var Logger *log.Logger
var Store *sessions.CookieStore
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

	Logger = log.New(os.Stdout, "logger: ", log.Lshortfile)

	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	authentication.Logger = Logger
	authentication.UserRepo = urepo.NewUserRepo(DB)
	authentication.Store = Store

	r := mux.NewRouter()
	r.HandleFunc("/login", authentication.LoginHandler)
	r.HandleFunc("/logout", authentication.Logout)
	r.HandleFunc("/callback", authentication.CallbackHandler)
	r.HandleFunc("/healthcheck", HealthcheckHandler)
	authedRouter := r.NewRoute().Subrouter()
	authedRouter.Use(ensureAuthed)
	authedRouter.HandleFunc("/", HomeHandler)

	authedRouter.HandleFunc("/search", SearchHandler)
	authedRouter.HandleFunc("/note", AddNoteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/delete", DeleteNoteHandler)
	authedRouter.HandleFunc("/api/done", ApiToggleNoteHandler)
	authedRouter.HandleFunc("/api/note/{id:[0-9]+}", ApiEditNoteHandler).Methods("PUT")
	authedRouter.HandleFunc("/api/notes", ApiNotesHandler).Methods("GET")
	authedRouter.HandleFunc("/api/todos", ApiTodosHandler).Methods("GET")
	authedRouter.HandleFunc("/api/readings", ApiReadingHandler).Methods("GET")

	authedRouter.PathPrefix("/public/").HandlerFunc(serveResources)

  // Catchall router
  r.PathPrefix("/").HandlerFunc(HomeHandler)

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         "0.0.0.0:" + env.GetEnvWithFallback("PORT", "8080"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	Logger.Println("Server listening")
	log.Fatal(srv.ListenAndServe())
}

type PageData struct {
	Notes       []repo.Note
	JsonNotes   string
	PreviousDay string
	NextDay     string
	CurrentDay  string
}

type IsoDate struct {
	t time.Time
}

func newIsoDateFromString(isoString string) (*IsoDate, error) {
	parsedTime, err := time.Parse(time.RFC3339, isoString+"T00:00:00+11:00")

	if err != nil {
		return nil, err
	}

	return &IsoDate{parsedTime}, nil
}

func newIsoDate() *IsoDate {
	return &IsoDate{time.Now()}
}

func (isoDate *IsoDate) stringify() string {
	return fmt.Sprintf("%d-%02d-%02d", isoDate.t.Year(), int(isoDate.t.Month()), isoDate.t.Day())
}

func (isoDate *IsoDate) timify() time.Time {
	return isoDate.t
}

func (isoDate *IsoDate) nextDay() *IsoDate {
	return &IsoDate{isoDate.t.AddDate(0, 0, 1)}
}

func (isoDate *IsoDate) previousDay() *IsoDate {
	return &IsoDate{isoDate.t.AddDate(0, 0, -1)}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date, ok := vars["date"]
	t := newIsoDate()
	var nextDay *IsoDate

	if ok {
		isoDate, err := newIsoDateFromString(date)

		if err != nil {
			handleUnexpectedError(w, err)
			return
		}

		t = isoDate
	}
	nextDay = t.nextDay()
	previousDay := t.previousDay()

	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetAllSince(r.Context(), t.t)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	jn, _ := json.Marshal(notes)
	pageData := PageData{
		JsonNotes:   string(jn),
		PreviousDay: previousDay.stringify(),
		NextDay:     nextDay.stringify(),
		CurrentDay:  t.stringify(),
	}

	RenderTemplate(w, "home", pageData)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, "todo", PageData{})
}
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ViewNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteId := vars["noteId"]

	noteRepo := repo.NewNoteRepo(DB, Logger)
	note, err := noteRepo.Get(r.Context(), repo.NoteID(noteId))

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: []repo.Note{note}}
	RenderTemplate(w, "view", pageData)
}

type DoneApiPayload struct {
	Id string `json:"id"`
}

func ApiToggleNoteHandler(w http.ResponseWriter, r *http.Request) {
	var payload DoneApiPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	noteRepo := repo.NewNoteRepo(DB, Logger)
	err = noteRepo.ToggleDone(r.Context(), repo.NoteID(payload.Id))

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ApiReadingHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetByTag(r.Context(), "to read")

	if err != nil {
		Logger.Println(err.Error())
		handleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func ApiTodosHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetByTag(r.Context(), "todo")

	if err != nil {
		Logger.Println(err.Error())
		handleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func ApiNotesHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	from := r.FormValue("from")
	to := r.FormValue("to")
	fromTime, err := newIsoDateFromString(from)
	if err != nil {
		Logger.Println(err.Error())
		handleUnexpectedError(w, err)
		return
	}
	toTime, err := newIsoDateFromString(to)
	if err != nil {
		Logger.Println(err.Error())
		handleUnexpectedError(w, err)
		return
	}

	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetAllBetween(r.Context(), fromTime.timify(), toTime.timify())

	if err != nil {
		Logger.Println(err.Error())
		handleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

type EditApiPayload struct {
	Id   string `json:"id"`
	Body string `json:"body"`
	Tags string `json:"tags"`
}

func ApiEditNoteHandler(w http.ResponseWriter, r *http.Request) {
	var payload EditApiPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	noteRepo := repo.NewNoteRepo(DB, Logger)
	err = noteRepo.Edit(r.Context(), repo.NoteID(payload.Id), payload.Body, payload.Tags)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	noteRepo := repo.NewNoteRepo(DB, Logger)
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

	noteRepo := repo.NewNoteRepo(DB, Logger)
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

	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.Search(r.Context(), query)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	jn, _ := json.Marshal(notes)
	pageData := PageData{JsonNotes: string(jn)}

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
}

func getUserFromSession(r *http.Request) (urepo.User, bool) {
	sessionState, err := Store.Get(r, "auth")
	if err != nil {
		println(err.Error())
	}
	userRepo := urepo.NewUserRepo(DB)
	userId, ok := sessionState.Values["user_id"].(string)

	if ok {
		user, _ := userRepo.Get(r.Context(), urepo.Auth0ID(userId))
		return user, true
	} else {
		return urepo.User{}, false
	}
}

func ensureAuthed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := getUserFromSession(r)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
	})
}
