package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/thrgamon/go-utils/env"
	urepo "github.com/thrgamon/go-utils/repo/user"
	"github.com/thrgamon/go-utils/web/authentication"
	"github.com/thrgamon/nous/database"
	"github.com/thrgamon/nous/environment"
	isoDate "github.com/thrgamon/nous/iso_date"
	"github.com/thrgamon/nous/logger"
	"github.com/thrgamon/nous/notes"
	"github.com/thrgamon/nous/repo"
	"github.com/thrgamon/nous/templates"
	"github.com/thrgamon/nous/web"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	Store *sessions.CookieStore
)

func init() {
	var e environment.Environment
	if env.GetEnvWithFallback("ENV", "production") == "development" {
		e = environment.Development
	} else {
		e = environment.Production
	}

	database.Init()
	logger.Init()
	templates.Init(e)

	Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	authentication.Logger = logger.Logger
	authentication.UserRepo = urepo.NewUserRepo(database.Database)
	authentication.Store = Store
}

func main() {
	defer database.Database.Close()
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

	authedRouter.HandleFunc("/note", notes.CreateNoteHandler).Methods("POST")
	authedRouter.HandleFunc("/note/{id:[0-9]+}", notes.ViewNoteHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/delete", notes.DeleteNoteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", notes.EditNoteHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", notes.UpdateNoteHandler).Methods("PUT")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/toggle", notes.ToggleNoteHandler)
	authedRouter.HandleFunc("/api/done", notes.ApiToggleNoteHandler)
	authedRouter.HandleFunc("/api/note/{id:[0-9]+}", notes.ApiEditNoteHandler).Methods("PUT")
	authedRouter.HandleFunc("/api/notes", notes.ApiNotesHandler).Methods("GET")
	authedRouter.HandleFunc("/api/todos", ApiTodosHandler).Methods("GET")
	authedRouter.HandleFunc("/api/readings", ApiReadingHandler).Methods("GET")

	authedRouter.PathPrefix("/public/").HandlerFunc(web.ServeResources)

	// Catchall router
	r.PathPrefix("/").HandlerFunc(HomeHandler)

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         "0.0.0.0:" + env.GetEnvWithFallback("PORT", "8080"),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Logger.Println("Server listening")
	logger.Logger.Fatal(srv.ListenAndServe())
}

type PageData struct {
	Notes       []repo.Note
	JsonNotes   string
	PreviousDay string
	NextDay     string
	CurrentDay  string
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(database.Database, logger.Logger)
	notes, err := noteRepo.GetByTag(r.Context(), "todo")

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{
		Notes: notes,
	}

	templates.RenderTemplate(w, "test", pageData)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date, ok := vars["date"]
	t := isoDate.NewIsoDate()
	var nextDay *isoDate.IsoDate

	if ok {
		isoDate, err := isoDate.NewIsoDateFromString(date)

		if err != nil {
			web.HandleUnexpectedError(w, err)
			return
		}

		t = isoDate
	}
	nextDay = t.NextDay()
	previousDay := t.PreviousDay()

	noteRepo := repo.NewNoteRepo(database.Database, logger.Logger)
	notes, err := noteRepo.GetAllSince(r.Context(), t.Time)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{
		Notes:       notes,
		PreviousDay: previousDay.Stringify(),
		NextDay:     nextDay.Stringify(),
		CurrentDay:  t.Stringify(),
	}

	templates.RenderTemplate(w, "home", pageData)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	templates.RenderTemplate(w, "todo", PageData{})
}
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ApiReadingHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(database.Database, logger.Logger)
	notes, err := noteRepo.GetByTag(r.Context(), "to read")

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func ApiTodosHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo(database.Database, logger.Logger)
	notes, err := noteRepo.GetByTag(r.Context(), "todo")

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func TagHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tag := r.FormValue("tag")

	noteRepo := repo.NewNoteRepo(database.Database, logger.Logger)
	notes, err := noteRepo.GetByTag(r.Context(), tag)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	templates.RenderTemplate(w, "home", pageData)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	query := r.FormValue("query")

	noteRepo := repo.NewNoteRepo(database.Database, logger.Logger)
	notes, err := noteRepo.Search(r.Context(), query)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	templates.RenderTemplate(w, "home", pageData)
}
