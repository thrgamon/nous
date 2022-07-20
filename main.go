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
	web.Init()

	authentication.Logger = logger.Logger
	authentication.UserRepo = urepo.NewUserRepo(database.Database)
	authentication.Store = web.Store
}

func main() {
	defer database.Database.Close()
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/login", authentication.LoginHandler)
	r.HandleFunc("/logout", authentication.Logout)
	r.HandleFunc("/callback", authentication.CallbackHandler)
	r.HandleFunc("/healthcheck", HealthcheckHandler)

	authedRouter := r.NewRoute().Subrouter()
	authedRouter.Use(web.EnsureAuthed)

	authedRouter.HandleFunc("/t/{date}", HomeHandler)
	authedRouter.HandleFunc("/review", ReviewHandler)
	authedRouter.HandleFunc("/search", SearchHandler)
	authedRouter.HandleFunc("/tag", TagHandler)

	authedRouter.HandleFunc("/note", notes.CreateHandler).Methods("POST")
	authedRouter.HandleFunc("/note/{id:[0-9]+}", notes.ViewNoteHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/delete", notes.DeleteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", notes.EditHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", notes.UpdateHandler).Methods("PUT")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/toggle", notes.ToggleHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/review", notes.ReviewedHandler).Methods("PATCH")
	authedRouter.HandleFunc("/api/todos", ApiTodosHandler).Methods("GET")
	authedRouter.HandleFunc("/api/readings", ApiReadingHandler).Methods("GET")

	authedRouter.PathPrefix("/public/").HandlerFunc(web.ServeResources)

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

	noteRepo := repo.NewNoteRepo()
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

func ReviewHandler(w http.ResponseWriter, r *http.Request) {
	notes, err := repo.NewNoteRepo().GetForReview(r.Context())

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{
		Notes: notes,
	}

	templates.RenderTemplate(w, "review", pageData)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	templates.RenderTemplate(w, "todo", PageData{})
}
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ApiReadingHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := repo.NewNoteRepo()
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
	noteRepo := repo.NewNoteRepo()
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

	noteRepo := repo.NewNoteRepo()
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

	noteRepo := repo.NewNoteRepo()
	notes, err := noteRepo.Search(r.Context(), query)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	templates.RenderTemplate(w, "home", pageData)
}
