package notes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	isoDate "github.com/thrgamon/nous/iso_date"
	"github.com/thrgamon/nous/logger"
	"github.com/thrgamon/nous/repo"
	"github.com/thrgamon/nous/templates"
	"github.com/thrgamon/nous/web"
)


func ViewNoteHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	noteRepo := repo.NewNoteRepo()
	note, err := noteRepo.Get(r.Context(), repo.NoteID(id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_toggle", note)
}

type DoneApiPayload struct {
	Id string `json:"id"`
}

func ToggleNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	noteRepo := repo.NewNoteRepo()
	_, err := noteRepo.ToggleDone(r.Context(), repo.NoteID(id))
	note, err := noteRepo.Get(r.Context(), repo.NoteID(id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_toggle", note)
}

func EditNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	noteRepo := repo.NewNoteRepo()
	note, err := noteRepo.Get(r.Context(), repo.NoteID(id))

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_edit", note)
}

func UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	r.ParseForm()

	body := r.FormValue("body")
	tags := r.FormValue("tags")

	noteRepo := repo.NewNoteRepo()
	err := noteRepo.Edit(r.Context(), repo.NoteID(id), body, tags)
	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	note, err := noteRepo.Get(r.Context(), repo.NoteID(id))
	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_toggle", note)
}

func ApiToggleNoteHandler(w http.ResponseWriter, r *http.Request) {
	var payload DoneApiPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	noteRepo := repo.NewNoteRepo()
	_, err = noteRepo.ToggleDone(r.Context(), repo.NoteID(payload.Id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ApiNotesHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	from := r.FormValue("from")
	to := r.FormValue("to")
	fromTime, err := isoDate.NewIsoDateFromString(from)
	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}
	toTime, err := isoDate.NewIsoDateFromString(to)
	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	noteRepo := repo.NewNoteRepo()
	notes, err := noteRepo.GetAllBetween(r.Context(), fromTime.Timify(), toTime.Timify())

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
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

	noteRepo := repo.NewNoteRepo()
	err = noteRepo.Edit(r.Context(), repo.NoteID(payload.Id), payload.Body, payload.Tags)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	noteRepo := repo.NewNoteRepo()
	err := noteRepo.Delete(r.Context(), repo.NoteID(id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	body := r.FormValue("body")
	tags := r.FormValue("tags")

	noteRepo := repo.NewNoteRepo()
	err := noteRepo.Add(r.Context(), body, tags)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
