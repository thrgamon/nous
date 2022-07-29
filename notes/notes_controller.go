package notes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thrgamon/nous/logger"
	"github.com/thrgamon/nous/templates"
	"github.com/thrgamon/nous/web"
)

func ViewNoteHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	noteRepo := NewNoteRepo()
	note, err := noteRepo.Get(r.Context(), NoteID(id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_note", note)
}

func ToggleHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	noteRepo := NewNoteRepo()
	_, err := noteRepo.ToggleDone(r.Context(), NoteID(id))
	note, err := noteRepo.Get(r.Context(), NoteID(id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_note", note)
}

func ReviewedHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	noteRepo := NewNoteRepo()
	if err := noteRepo.MarkReviewed(r.Context(), NoteID(id)); err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	w.WriteHeader(200)
}

func EditHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	noteRepo := NewNoteRepo()
	note, err := noteRepo.Get(r.Context(), NoteID(id))

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_edit", note)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	r.ParseForm()

	body := r.FormValue("body")
	tags := r.FormValue("tags")

	noteRepo := NewNoteRepo()
	err := noteRepo.Edit(r.Context(), NoteID(id), body, tags)
	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	note, err := noteRepo.Get(r.Context(), NoteID(id))
	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	templates.RenderTemplate(w, "_note", note)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	noteRepo := NewNoteRepo()
	err := noteRepo.Delete(r.Context(), NoteID(id))

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	body := r.FormValue("body")
	tags := r.FormValue("tags")

	noteRepo := NewNoteRepo()
	err := noteRepo.Add(r.Context(), body, tags)

	if err != nil {
		web.HandleUnexpectedError(w, err)
		return
	}

	w.Header().Add("HX-Refresh", "true")
}
