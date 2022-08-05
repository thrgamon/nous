package notes

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thrgamon/nous/database"
	"github.com/thrgamon/nous/logger"
	"github.com/thrgamon/nous/url"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type NoteID string

type Note struct {
	ID   NoteID   `json:"id"`
	Body string   `json:"body"`
	Tags []string `json:"tags"`
	Done bool     `json:"done"`

	DisplayBody template.HTML
	DisplayTags string
}

type NoteRepo struct {
	db     *pgxpool.Pool
	logger *log.Logger
}

func NewNoteRepo() *NoteRepo {
	db := database.Database
	logger := logger.Logger
	return &NoteRepo{db: db, logger: logger}
}

func (rr NoteRepo) Get(ctx context.Context, id NoteID) (Note, error) {
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
      note_search.id = $1;`,
		id,
	)

	defer rows.Close()

	if err != nil {
		rr.logger.Println(err.Error())
		return Note{}, err
	}

	notes, err := rr.parseData(rows)
	if err != nil {
		rr.logger.Println(err.Error())
		return Note{}, err
	}

	return notes[0], nil
}

func (rr NoteRepo) GetAllSince(ctx context.Context, t time.Time) ([]Note, error) {
	return rr.GetAllBetween(ctx, startOfDay(t), endOfDay(t))
}

func (rr NoteRepo) GetByTags(ctx context.Context, tags string) ([]Note, error) {
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
    string_to_array($1, ',') <@ tags::text[] AND done=false
    ORDER BY
      note_search.id DESC`,
		tags,
	)

	defer rows.Close()

	if err != nil {
		return notes, err
	}

	return rr.parseData(rows)
}

func (rr NoteRepo) parseData(rows pgx.Rows) ([]Note, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	var notes []Note
	var err error

	for rows.Next() {
		var buf bytes.Buffer
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

		if err := md.Convert([]byte(body), &buf); err != nil {
			panic(err)
		}

		notes = append(
			notes,
			Note{
				ID:   NoteID(fmt.Sprint(id)),
				Body: body,
				Tags: tags,
				Done: done,

				DisplayBody: template.HTML(buf.String()),
				DisplayTags: strings.Join(tags, ", "),
			},
		)
	}

	if err != nil {
		rr.logger.Println(err.Error())
		return notes, err
	}

	return notes, nil
}

func (rr NoteRepo) GetForReview(ctx context.Context) ([]Note, error) {
	var notes []Note
	rows, err := rr.db.Query(
		ctx,
		`SELECT
      note_search.id,
      body,
      tags,
      done,
      note_search.inserted_at
    FROM
      note_search
    WHERE
      reviewed_at IS NULL
    ORDER BY
      note_search.id DESC`,
	)
	defer rows.Close()

	if err != nil {
		rr.logger.Println(err.Error())
		return notes, err
	}

	return rr.parseData(rows)
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
		rr.logger.Println(err.Error())
		return notes, err
	}

	return rr.parseData(rows)
}

func (rr NoteRepo) ToggleDone(ctx context.Context, noteId NoteID) (bool, error) {
	var done bool
	err := rr.db.QueryRow(ctx, "UPDATE notes SET done = NOT done WHERE id = $1 RETURNING done", noteId).Scan(&done)
	return done, err
}

func (rr NoteRepo) MarkReviewed(ctx context.Context, noteId NoteID) error {
	_, err := rr.db.Exec(ctx, "UPDATE notes SET reviewed_at = NOW() WHERE id = $1", noteId)
	return err
}

func (rr NoteRepo) Delete(ctx context.Context, noteId NoteID) error {
	error := rr.withTransaction(ctx, func() error {
		_, err := rr.db.Exec(ctx, "DELETE FROM tags WHERE note_id = $1", noteId)
		if err != nil {
			rr.logger.Println(err.Error())
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
			rr.logger.Println(err.Error())
			return err
		}

		combinedTags := assembleTags(body, tags)

		if len(combinedTags) > 0 {
			for _, string := range combinedTags {
        var tagId int
				fmtString := strings.TrimSpace(strings.ToLower(string))

        err := rr.db.QueryRow(ctx, "INSERT INTO tags (tag) VALUES ($1) ON CONFLICT (tag) DO UPDATE SET updated_at = NOW() RETURNING id", fmtString).Scan(&tagId)
        if err != nil {return err}

        _, err = rr.db.Exec(ctx, "INSERT INTO notetags (tag_id, note_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", tagId, noteId)
        if err != nil {return err}
			}

			return err
		}
		return nil
	})

	go url.ExtractURLMetadata(body)

	return error
}

func (rr NoteRepo) AddTag(ctx context.Context, noteId NoteID, tag string) error {
	_, err := rr.db.Exec(ctx, "INSERT INTO tags (note_id, tag) VALUES ($1, $2)", noteId, tag)
	return err
}

func (rr NoteRepo) Edit(ctx context.Context, noteId NoteID, body string, tags string) error {
	error := rr.withTransaction(ctx, func() error {
		_, err := rr.db.Exec(ctx, "UPDATE notes SET body=$1 WHERE notes.id = $2", body, noteId)

		if err != nil {
			rr.logger.Println(err.Error())
			return err
		}

		_, err = rr.db.Exec(ctx, "DELETE FROM notetags WHERE note_id = $1", noteId)

		if err != nil {
			rr.logger.Println(err.Error())
			return err
		}

		combinedTags := assembleTags(body, tags)

		if len(combinedTags) > 0 {
			for _, string := range combinedTags {
        var tagId int
				fmtString := strings.TrimSpace(strings.ToLower(string))
        err := rr.db.QueryRow(ctx, "INSERT INTO tags (tag) VALUES ($1)", fmtString).Scan(&tagId)
        if err != nil {return err}
        _, err = rr.db.Exec(ctx, "INSERT INTO notetags (tag_id, note_id) VALUES ($1, $2)", tagId, noteId)
        if err != nil {return err}
			}

			return err
		}
		return nil
	})

	go url.ExtractURLMetadata(body)

	return error
}

func assembleTags(body string, tags string) []string {
	var mainTags []string
	if tags != "" {
		for _, tag := range strings.Split(strings.TrimSpace(tags), ",") {
			if tag != "" {
				mainTags = append(mainTags, tag)
			}
		}
	}
	peopleTags := ExtractPeople(body)
	return append(mainTags, peopleTags...)
}

func (rr NoteRepo) GetTodos(ctx context.Context) ([]Note, error) {
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
  		('todo' = ANY(tags) OR body LIKE '%- [ ]%') AND done=false
    ORDER BY
      note_search.id DESC`,
	)
	defer rows.Close()

	if err != nil {
		rr.logger.Println(err.Error())
		return notes, err
	}

	return rr.parseData(rows)
}

func (rr NoteRepo) Search(ctx context.Context, searchQuery string) ([]Note, error) {
	var notes []Note

	tsquery := strings.Join(strings.Split(searchQuery, " "), " | ")

	// Using a subtable so we can order by rank without
	// returning it
	rows, err := rr.db.Query(
		ctx,
		`SELECT
	id,
	body,
	tags,
	done,
	inserted_at
FROM (
	SELECT
		note_search.id AS id,
		body,
		tags,
		done,
		inserted_at,
		ts_rank(note_search.doc, to_tsquery($1)) AS rank
	FROM
		note_search
	WHERE
		note_search.doc @@ to_tsquery($1) AND done = false
	ORDER BY
		rank DESC,
		inserted_at DESC) subtable`,
		tsquery,
	)
	defer rows.Close()

	if err != nil {
		rr.logger.Println(err.Error())
		return notes, err
	}

	return rr.parseData(rows)
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
