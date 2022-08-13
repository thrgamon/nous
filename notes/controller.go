package notes

import (
	"net/http"
	"strconv"
	"strings"

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

type StatusPageData struct {
	Statuses []StatusNotes
}

type StatusNotes struct {
	Name  string
	Notes []Note
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	r.ParseForm()
	priorityLevel, err := strconv.Atoi(r.FormValue("priority"))
	if err != nil {
		panic(err)
	}

	noteRepo := NewNoteRepo()

	if err := noteRepo.SetPriority(r.Context(), NoteID(id), GetPriorityLevel(priorityLevel)); err != nil {
		panic(err)
	}

	pageData := StatusPageData{Statuses: []StatusNotes{}}

	notes, err := noteRepo.GetByPriority(r.Context())
	if err != nil {
		panic(err)
	}

	m := make(map[PriorityLevel][]Note)

	for _, note := range notes {
		arr, ok := m[note.Priority]
		if ok {
			arr = append(arr, note)
		} else {
			arr = []Note{note}
		}
		m[note.Priority] = arr
	}

	statuses := []PriorityLevel{Unprioritised, ImportantAndUrgent, Important, Urgent, Someday}
	for _, status := range statuses {
		statusNote := StatusNotes{Name: string(status), Notes: m[status]}
		pageData.Statuses = append(pageData.Statuses, statusNote)
	}

	templates.RenderTemplate(w, "_todos", pageData)
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

func ToggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	todoIndex := vars["todoIndex"]

	noteRepo := NewNoteRepo()
	note, err := noteRepo.Get(r.Context(), NoteID(id))
	ti, _ := strconv.Atoi(todoIndex)
	newBody, _ := ToggleTodo(note.Body, ti)
	err = noteRepo.Edit(r.Context(), NoteID(id), newBody, strings.Join(note.Tags, ","))

	if err != nil {
		logger.Logger.Println(err.Error())
		web.HandleUnexpectedError(w, err)
		return
	}

	note, _ = noteRepo.Get(r.Context(), NoteID(id))

	templates.RenderTemplate(w, "_note", note)
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
