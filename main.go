package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/thrgamon/go-utils/env"
	urepo "github.com/thrgamon/go-utils/repo/user"
	"github.com/thrgamon/go-utils/web/authentication"
	isoDate "github.com/thrgamon/nous/iso_date"
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

var (
	DB        *pgxpool.Pool
	Templates map[string]*template.Template
	Logger    *log.Logger
	Store     *sessions.CookieStore
	ENV       Environment
)

func init() {
	if env.GetEnvWithFallback("ENV", "production") == "development" {
		ENV = Development
	} else {
		ENV = Production
	}

	DB = initDB()

	Templates = cacheTemplates()

	Logger = log.New(os.Stdout, "logger: ", log.Lshortfile)

	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	authentication.Logger = Logger
	authentication.UserRepo = urepo.NewUserRepo(DB)
	authentication.Store = Store
}

func main() {
	defer DB.Close()
	r := mux.NewRouter()
	r.HandleFunc("/login", authentication.LoginHandler)
	r.HandleFunc("/logout", authentication.Logout)
	r.HandleFunc("/callback", authentication.CallbackHandler)
	r.HandleFunc("/healthcheck", HealthcheckHandler)

	authedRouter := r.NewRoute().Subrouter()
	authedRouter.Use(ensureAuthed)

	authedRouter.HandleFunc("/t/{date}", HomeHandler)
	authedRouter.HandleFunc("/search", SearchHandler)
	authedRouter.HandleFunc("/tag", TagHandler)
	authedRouter.HandleFunc("/note", AddNoteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}", NoteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/delete", DeleteNoteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", EditNoteHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", UpdateNoteHandler).Methods("PUT")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/toggle", ToggleNoteHandler)
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

func TestHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetByTag(r.Context(), "todo")

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{
		Notes: notes,
	}

	RenderTemplate(w, "test", pageData)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date, ok := vars["date"]
	t := isoDate.NewIsoDate()
	var nextDay *isoDate.IsoDate

	if ok {
		isoDate, err := isoDate.NewIsoDateFromString(date)

		if err != nil {
			handleUnexpectedError(w, err)
			return
		}

		t = isoDate
	}
	nextDay = t.NextDay()
	previousDay := t.PreviousDay()

	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetAllSince(r.Context(), t.Time)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{
		Notes:       notes,
		PreviousDay: previousDay.Stringify(),
		NextDay:     nextDay.Stringify(),
		CurrentDay:  t.Stringify(),
	}

	RenderTemplate(w, "home", pageData)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, "todo", PageData{})
}
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
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

func TagHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tag := r.FormValue("tag")

	noteRepo := repo.NewNoteRepo(DB, Logger)
	notes, err := noteRepo.GetByTag(r.Context(), tag)

	if err != nil {
		handleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	RenderTemplate(w, "home", pageData)
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

	pageData := PageData{Notes: notes}

	RenderTemplate(w, "home", pageData)
}
