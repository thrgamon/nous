package repo

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/jackc/pgx/v4/pgxpool"
)

type NoteID string

type Note struct {
	ID   NoteID
	Body template.HTML
	Tags []string
	Done bool
}

type NoteRepo struct {
	storage map[uint]Note
	db      *pgxpool.Pool
}

func NewNoteRepo(db *pgxpool.Pool) *NoteRepo {
	var repo NoteRepo
	repo.storage = make(map[uint]Note)
	repo.db = db
	return &repo
}

func (rr NoteRepo) Get(ctx context.Context, id NoteID) (Note, error) {
	var body string
	var tags []string
	err := rr.db.QueryRow(
		ctx,
		`SELECT
      body,
      tags
    FROM
      note_search
    WHERE
      note_search.id = $1;`,
		id,
	).Scan(&body, &tags)

	if err != nil {
		return Note{}, err
	}

	if err != nil {
		return Note{}, err
	}

	note := Note{
		ID:   id,
		Body: template.HTML(markdown.ToHTML([]byte(body), nil, nil)),
		Tags: tags,
	}

	return note, nil
}

func (rr NoteRepo) GetAll(ctx context.Context) ([]Note, error) {
	var notes []Note

	rows, err := rr.db.Query(
		ctx,
		`SELECT
      note_search.id,
      body,
      tags,
      done
    FROM
      note_search
    ORDER BY
      note_search.id DESC`,
	)
	defer rows.Close()

	if err != nil {
		return notes, err
	}

	for rows.Next() {
		var id int
		var body string
		var tags []string
		var done bool
    err := rows.Scan(&id, &body, &tags, &done)

		if err != nil {
			return notes, err
		}

		notes = append(
			notes,
			Note{
				ID:   NoteID(fmt.Sprint(id)),
				Body: template.HTML(markdown.ToHTML([]byte(body), nil, nil)),
				Tags: tags,
				Done: done,
			},
		)
	}

	if err != nil {
		log.Fatal(err.Error())
		return notes, err
	}

	return notes, nil
}

func (rr NoteRepo) ToggleDone(ctx context.Context, noteId NoteID) error {
	_, err := rr.db.Exec(ctx, "UPDATE notes SET done = NOT done WHERE id = $1", noteId)

	return err
}
func (rr NoteRepo) Add(ctx context.Context, body string, tags string) error {
	error := rr.withTransaction(ctx, func() error {
		var noteId int
		err := rr.db.QueryRow(ctx, "INSERT INTO notes (body) VALUES ($1) RETURNING id", body).Scan(&noteId)

    if tags != "" {
      splitTags := strings.Split(tags, " ")

      for _, string := range splitTags {
        fmtString := strings.TrimSpace(strings.ToLower(string))
        rr.db.Exec(ctx, "INSERT INTO tags (note_id, tag) VALUES ($1, $2)", noteId, fmtString)
      }
    }

		return err
	})

	return error
}

func (rr NoteRepo) Search(ctx context.Context, searchQuery string) ([]Note, error) {
	var notes []Note

	tsquery := strings.Join(strings.Split(searchQuery, " "), " | ")

	rows, err := rr.db.Query(
		ctx,
		`SELECT
      note_search.id,
      body,
      tags,
      done
    FROM
      note_search
    WHERE
       note_search.doc @@ to_tsquery($1)
    ORDER BY
      note_search.id DESC`,
		tsquery,
	)
	defer rows.Close()

	if err != nil {
		return notes, err
	}

	for rows.Next() {
		var id int
		var body string
		var tags []string
		var done bool
    err := rows.Scan(&id, &body, &tags, &done)

		if err != nil {
			return notes, err
		}

		notes = append(
			notes,
			Note{
				ID:   NoteID(fmt.Sprint(id)),
				Body: template.HTML(markdown.ToHTML([]byte(body), nil, nil)),
				Tags: tags,
        Done: done,
			},
		)
	}

	if err != nil {
		return notes, err
	}

	return notes, nil
}

func (rr NoteRepo) withTransaction(ctx context.Context, fn func() error) error {
	tx, err := rr.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = fn()
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
