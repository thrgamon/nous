package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/thrgamon/go-utils/env"
	urepo "github.com/thrgamon/go-utils/repo/user"
	"github.com/thrgamon/go-utils/web/authentication"
	"github.com/thrgamon/nous/api"
	"github.com/thrgamon/nous/contexts"
	"github.com/thrgamon/nous/database"
	"github.com/thrgamon/nous/environment"
	isoDate "github.com/thrgamon/nous/iso_date"
	"github.com/thrgamon/nous/logger"
	"github.com/thrgamon/nous/notes"
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
	r.HandleFunc("/login", authentication.LoginHandler)
	r.HandleFunc("/logout", authentication.Logout)
	r.HandleFunc("/callback", authentication.CallbackHandler)
	r.HandleFunc("/healthcheck", HealthcheckHandler)

	authedRouter := r.NewRoute().Subrouter()
	authedRouter.Use(web.EnsureAuthed)

	authedRouter.HandleFunc("/", HomeHandler)
	authedRouter.HandleFunc("/v2", V2Handler)
	authedRouter.HandleFunc("/t/{date}", HomeHandler)
	authedRouter.HandleFunc("/review", ReviewHandler)
	authedRouter.HandleFunc("/search", SearchHandler)
	authedRouter.HandleFunc("/live_search", LiveSearchHandler)
	authedRouter.HandleFunc("/tag", TagHandler)

	authedRouter.HandleFunc("/active-context", GetActiveContextHandler).Methods("GET")
	authedRouter.HandleFunc("/switch-context", GetContextHandler).Methods("GET")
	authedRouter.HandleFunc("/switch-context/{context:[a-z]+}", UpdateContextHandler).Methods("PUT")
	authedRouter.HandleFunc("/note", notes.CreateHandler).Methods("POST")
	authedRouter.HandleFunc("/note/{id:[0-9]+}", notes.ViewNoteHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/delete", notes.DeleteHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", notes.EditHandler).Methods("GET")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/edit", notes.UpdateHandler).Methods("PUT")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/toggle", notes.ToggleHandler)
	authedRouter.HandleFunc("/note/{id:[0-9]+}/review", notes.ReviewedHandler).Methods("PATCH")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/status", notes.StatusHandler).Methods("PATCH")
	authedRouter.HandleFunc("/note/{id:[0-9]+}/todo/{todoIndex:[0-9]+}", notes.ToggleTodoHandler).Methods("PUT")
	authedRouter.HandleFunc("/todos", TodoHandler).Methods("GET")
	authedRouter.HandleFunc("/api/readings", ApiReadingHandler).Methods("GET")

	authedRouter.HandleFunc("/api/notes", api.AllNotes).Methods("GET")
	authedRouter.HandleFunc("/api/note", api.CreateNote).Methods("POST")
	authedRouter.HandleFunc("/api/note/{id:[0-9]+}", api.DeleteNote).Methods("DELETE")
	authedRouter.HandleFunc("/api/note/{id:[0-9]+}", api.EditNote).Methods("PUT")

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
	Notes       []notes.Note
	JsonNotes   string
	PreviousDay string
	NextDay     string
	CurrentDay  string
	Context     string
}

func V2Handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/v2.html")
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

	noteRepo := notes.NewNoteRepo()
	notes, err := noteRepo.GetAllSince(r.Context(), t.Time)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	context := contexts.NewContextRepo().GetActiveContext(r.Context())

	pageData := PageData{
		Notes:       notes,
		PreviousDay: previousDay.Stringify(),
		NextDay:     nextDay.Stringify(),
		CurrentDay:  t.Stringify(),
		Context:     context,
	}

	templates.RenderTemplate(w, "home", pageData)
}

func ReviewHandler(w http.ResponseWriter, r *http.Request) {
	notes, err := notes.NewNoteRepo().GetForReview(r.Context())

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{
		Notes: notes,
	}

	templates.RenderTemplate(w, "review", pageData)
}

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ApiReadingHandler(w http.ResponseWriter, r *http.Request) {
	noteRepo := notes.NewNoteRepo()
	notes, err := noteRepo.GetByTags(r.Context(), "to read")

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func TodoHandler(w http.ResponseWriter, r *http.Request) {
	context := contexts.NewContextRepo().GetActiveContext(r.Context())
	noteRepo := notes.NewNoteRepo()
	pageData := notes.StatusPageData{Statuses: []notes.StatusNotes{}, Context: context + ", todo"}

	nts, err := noteRepo.GetByPriority(r.Context())
	if err != nil {
		panic(err)
	}

	m := make(map[notes.PriorityLevel][]notes.Note)

	for _, note := range nts {
		arr, ok := m[note.Priority]
		if ok {
			arr = append(arr, note)
		} else {
			arr = []notes.Note{note}
		}
		m[note.Priority] = arr
	}

	statuses := []notes.PriorityLevel{notes.Unprioritised, notes.ImportantAndUrgent, notes.Important, notes.Urgent, notes.Someday}
	for _, status := range statuses {
		statusNote := notes.StatusNotes{Name: string(status), Notes: m[status]}
		pageData.Statuses = append(pageData.Statuses, statusNote)
	}

	templates.RenderTemplate(w, "todos", pageData)
}

func TagHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tags := r.FormValue("tags")

	noteRepo := notes.NewNoteRepo()
	notes, err := noteRepo.GetByTags(r.Context(), tags)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		panic(err)
	}

	pageData := PageData{Notes: notes}

	templates.RenderTemplate(w, "notes", pageData)
}

func LiveSearchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	query := r.FormValue("query")

	noteRepo := notes.NewNoteRepo()
	notes, err := noteRepo.Search(r.Context(), query)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	templates.RenderTemplate(w, "_notes", pageData)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	query := r.FormValue("query")

	noteRepo := notes.NewNoteRepo()
	notes, err := noteRepo.Search(r.Context(), query)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	pageData := PageData{Notes: notes}

	templates.RenderTemplate(w, "search", pageData)
}

func GetContextHandler(w http.ResponseWriter, r *http.Request) {
	contextRepo := contexts.NewContextRepo()
	contexts := contextRepo.GetContexts(r.Context())

	templates.RenderTemplate(w, "_switch-context", contexts)
}

func GetActiveContextHandler(w http.ResponseWriter, r *http.Request) {
	contextRepo := contexts.NewContextRepo()
	activeContext := contextRepo.GetActiveContext(r.Context())

	templates.RenderTemplate(w, "_active-context", activeContext)
}

func UpdateContextHandler(w http.ResponseWriter, r *http.Request) {
	context := mux.Vars(r)["context"]

	contextRepo := contexts.NewContextRepo()
	contextRepo.UpdateContext(r.Context(), context)

	w.Header().Set("HX-Refresh", "true")
}
