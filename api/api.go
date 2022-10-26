package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"net/http"
	"os"
)

type Note struct {
	ID   string         `db:"id" json:"id"`
	Body string         `db:"body" json:"body"`
	Tags pq.StringArray `db:"tags" json:"tags"`
}

func AllNotes(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	r.ParseForm()

	from := r.FormValue("from")
	to := r.FormValue("to")
	defer db.Close()
	if err != nil {
		panic(err)
	}
	notes := []Note{}
	sqlStatement := "SELECT notes.id, body, tags FROM notes JOIN note_search on notes.id = note_search.id"
	order := " ORDER BY notes.id DESC"

	if to != "" || from != "" {
		sqlStatement += " WHERE"
	}

	if from != "" {
		sqlStatement += " inserted_at >= "
		sqlStatement += "'" + from + "'"
	}

	if to != "" {
		if from != "" {
			sqlStatement += " AND"
		}
		sqlStatement += " inserted_at <= "
		sqlStatement += "'" + to + "'"
	}

	sqlStatement += order

	db.Select(&notes, sqlStatement)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func CreateNote(w http.ResponseWriter, r *http.Request) {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(r.Body)
	var n Note
	err = dec.Decode(&n)

	if err != nil {
		panic(err)
	}

	var noteId string
	tx := db.MustBegin()
	err = tx.QueryRow("INSERT INTO notes (body) VALUES ($1) RETURNING id", n.Body).Scan(&noteId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

  insertTags(n.Tags, noteId, tx)

	tx.Commit()
	n.ID = noteId
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)
}

func insertTags(tags []string, noteId string, tx *sqlx.Tx) {
	var tagId string
	for _, tag := range tags {
		err := tx.QueryRow("INSERT INTO tags (tag) VALUES ($1) ON CONFLICT (tag) DO UPDATE SET updated_at = NOW() RETURNING id", tag).Scan(&tagId)
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		_, err = tx.Exec("INSERT INTO notetags (tag_id, note_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", tagId, noteId)

		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}
}

func DeleteNote(w http.ResponseWriter, r *http.Request) {
	noteId := mux.Vars(r)["id"]

	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()
	if err != nil {
		panic(err)
	}

	tx := db.MustBegin()
	_, err = tx.Exec("DELETE FROM notetags WHERE note_id = $1", noteId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	_, err = tx.Exec("DELETE FROM notes WHERE id = $1", noteId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	tx.Commit()
}

func EditNote(w http.ResponseWriter, r *http.Request) {
	noteId := mux.Vars(r)["id"]

	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(r.Body)
	var n Note
	err = dec.Decode(&n)

	if err != nil {
		panic(err)
	}

	tx := db.MustBegin()
	_, err = tx.Exec("UPDATE notes SET body=$1 WHERE notes.id = $2", n.Body, noteId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	_, err = tx.Exec("DELETE FROM notetags WHERE note_id = $1", noteId)
	if err != nil {
		tx.Rollback()
		panic(err)
	}

  insertTags(n.Tags, noteId, tx)

	tx.Commit()
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
