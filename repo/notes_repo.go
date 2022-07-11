package repo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type NoteID string

type Note struct {
	ID   NoteID   `json:"id"`
	Body string   `json:"body"`
	Tags []string `json:"tags"`
	Done bool     `json:"done"`
}

type NoteRepo struct {
	db     *pgxpool.Pool
	logger *log.Logger
}

func NewNoteRepo(db *pgxpool.Pool, logger *log.Logger) *NoteRepo {
	return &NoteRepo{db: db, logger: logger}
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

	note := Note{
		ID:   id,
		Body: body,
		Tags: tags,
	}

	return note, nil
}

func (rr NoteRepo) GetAllSince(ctx context.Context, t time.Time) ([]Note, error) {
	return rr.GetAllBetween(ctx, startOfDay(t), endOfDay(t))
}

func (rr NoteRepo) GetByTag(ctx context.Context, tag string) ([]Note, error) {
	var notes []Note
	rows, err := rr.db.Query(
		ctx,
		`SELECT
      note_search.id,
      body,
      tags,
      done,
      inserted_at
    FROM
      note_search
  	WHERE
  		$1 = ANY(tags) AND done = false
    ORDER BY
      note_search.id DESC`,
		tag,
	)

	defer rows.Close()

	if err != nil {
		return notes, err
	}

	return rr.parseData(rows)
}

func (rr NoteRepo) parseData(rows pgx.Rows) ([]Note, error) {
	var notes []Note
	var err error

	for rows.Next() {
		var id int
		var body string
		var tags []string
		var done bool
		var insertedAt time.Time
		err := rows.Scan(&id, &body, &tags, &done, &insertedAt)

		if err != nil {
			rr.logger.Println(err.Error())
			return notes, err
		}

		notes = append(
			notes,
			Note{
				ID:   NoteID(fmt.Sprint(id)),
				Body: body,
				Tags: tags,
				Done: done,
			},
		)
	}

	if err != nil {
		rr.logger.Println(err.Error())
		return notes, err
	}

	return notes, nil
}

func (rr NoteRepo) GetAllBetween(ctx context.Context, from time.Time, to time.Time) ([]Note, error) {
	var notes []Note
	rows, err := rr.db.Query(
		ctx,
		`SELECT
      note_search.id,
      body,
      tags,
      done,
      inserted_at
    FROM
      note_search
    WHERE
      inserted_at BETWEEN $1 AND $2
    ORDER BY
      note_search.id DESC`,
		from,
		to,
	)
	defer rows.Close()

	if err != nil {
		return notes, err
	}

	return rr.parseData(rows)
}

func (rr NoteRepo) ToggleDone(ctx context.Context, noteId NoteID) error {
	_, err := rr.db.Exec(ctx, "UPDATE notes SET done = NOT done WHERE id = $1", noteId)
	return err
}

func (rr NoteRepo) Delete(ctx context.Context, noteId NoteID) error {
	error := rr.withTransaction(ctx, func() error {
		_, err := rr.db.Exec(ctx, "DELETE FROM tags WHERE note_id = $1", noteId)
		if err != nil {
			return err
		}

		_, err = rr.db.Exec(ctx, "DELETE FROM notes WHERE id = $1", noteId)
		return err
	})
	return error
}

func (rr NoteRepo) Add(ctx context.Context, body string, tags string) error {
	error := rr.withTransaction(ctx, func() error {
		var noteId int
		err := rr.db.QueryRow(ctx, "INSERT INTO notes (body) VALUES ($1) RETURNING id", body).Scan(&noteId)

		if err != nil {
			return err
		}

		if tags != "" {
			splitTags := strings.Split(strings.TrimSpace(tags), ",")

			batch := &pgx.Batch{}

			for _, string := range splitTags {
				fmtString := strings.TrimSpace(strings.ToLower(string))
				batch.Queue("INSERT INTO tags (note_id, tag) VALUES ($1, $2)", noteId, fmtString)
			}

			br := rr.db.SendBatch(ctx, batch)
			err = br.Close()

			return err
		}
		return nil
	})

	return error
}

func (rr NoteRepo) Edit(ctx context.Context, noteId NoteID, body string, tags string) error {
	error := rr.withTransaction(ctx, func() error {
		_, err := rr.db.Exec(ctx, "UPDATE notes SET body=$1 WHERE notes.id = $2", body, noteId)

		if err != nil {
			return err
		}

		_, err = rr.db.Exec(ctx, "DELETE FROM tags WHERE note_id = $1", noteId)

		if err != nil {
			return err
		}

		if tags != "" {
			splitTags := strings.Split(strings.TrimSpace(tags), ",")

			batch := &pgx.Batch{}

			for _, string := range splitTags {
				fmtString := strings.TrimSpace(strings.ToLower(string))
				batch.Queue("INSERT INTO tags (note_id, tag) VALUES ($1, $2)", noteId, fmtString)
			}

			br := rr.db.SendBatch(ctx, batch)
			err = br.Close()

			return err
		}
		return nil
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
      done,
      inserted_at,
      ts_rank(note_search.doc, to_tsquery($1)) AS rank
    FROM
      note_search
    WHERE
       note_search.doc @@ to_tsquery($1)
    ORDER BY
      rank DESC, inserted_at DESC;`,
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
		err := rows.Scan(&id, &body, &tags, &done, nil, nil)

		if err != nil {
			return notes, err
		}

		notes = append(
			notes,
			Note{
				ID:   NoteID(fmt.Sprint(id)),
				Body: body,
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

func startOfDay(t time.Time) time.Time {
	melbourne, _ := time.LoadLocation("Australia/Melbourne")
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, melbourne)
}

func endOfDay(t time.Time) time.Time {
	melbourne, _ := time.LoadLocation("Australia/Melbourne")
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, melbourne)
}
