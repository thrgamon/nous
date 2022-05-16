package repo

import (
	"context"
	"log"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

type ResourceID uint

type Resource struct {
	ID    uint
	Link  url.URL
	Name  string
	Rank  int
	Voted bool
	Tags  []string
}

type ResourceRepo struct {
	storage map[uint]Resource
	db      *pgxpool.Pool
}

func NewResourceRepo(db *pgxpool.Pool) *ResourceRepo {
	var repo ResourceRepo
	repo.storage = make(map[uint]Resource)
	repo.db = db
	return &repo
}

func (rr ResourceRepo) Get(ctx context.Context, id uint, userId UserID) (Resource, error) {
	var link string
	var name string
	var rank int
	var voted int
	var tags []string
	err := rr.db.QueryRow(
		ctx,
		`SELECT
      link,
      name,
      rank,
      CASE WHEN votes.id IS NULL THEN 0 ELSE 1 END uservoted,
      tags
    FROM
      resource_search
      LEFT JOIN votes on votes.resource_id = resource_search.id AND votes.user_id = $1
    WHERE
      resource_search.id = $2
    ORDER BY
      rank DESC,
      inserted_at;`,
		userId,
		id,
	).Scan(&link, &name, &rank, &voted, &tags)

	if err != nil {
		return Resource{}, err
	}

	urlLink, err := url.Parse(link)

	if err != nil {
		return Resource{}, err
	}

	resource := Resource{ID: id, Link: *urlLink, Name: name, Rank: rank, Voted: voted == 1, Tags: tags}

	return resource, nil
}

func (rr ResourceRepo) GetAll(ctx context.Context, userId UserID) ([]Resource, error) {
	var resources []Resource

	rows, err := rr.db.Query(
		ctx,
		`SELECT
      resource_search.id,
      link,
      name,
      rank,
      CASE WHEN votes.id IS NULL THEN 0 ELSE 1 END uservoted,
      tags
    FROM
      resource_search
      LEFT JOIN votes on votes.resource_id = resource_search.id AND votes.user_id = $1
    ORDER BY
      rank DESC,
      inserted_at;`,
		userId,
	)
	defer rows.Close()

	if err != nil {
		return resources, err
	}

	for rows.Next() {
		var id int
		var link string
		var name string
		var rank int
		var voted int
		var tags []string
		rows.Scan(&id, &link, &name, &rank, &voted, &tags)

		urlLink, err := url.Parse(link)

		if err != nil {
			return resources, err
		}

		resources = append(
			resources,
			Resource{
				ID:    uint(id),
				Link:  *urlLink,
				Name:  name,
				Rank:  rank,
				Voted: voted == 1,
				Tags:  tags,
			},
		)
	}

	if err != nil {
		log.Fatal(err.Error())
		return resources, err
	}

	return resources, nil
}

func (rr ResourceRepo) Add(ctx context.Context, userId UserID, link string, name string, tags string) error {
	error := rr.withTransaction(ctx, func() error {
		var resourceId uint
		err := rr.db.QueryRow(ctx, "INSERT INTO resources (name, link, rank) VALUES ($1, $2, $3) RETURNING id", name, link, 0).Scan(&resourceId)

		splitTags := strings.Split(tags, ",")

		for _, string := range splitTags {
			fmtString := strings.TrimSpace(strings.ToLower(string))
			rr.db.Exec(ctx, "INSERT INTO tags (user_id, resource_id, tag) VALUES ($1, $2, $3)", userId, resourceId, fmtString)
		}

		return err
	})

	return error
}

func (rr ResourceRepo) Upvote(ctx context.Context, userId UserID, resourceId uint) error {
	_, err := rr.db.Exec(ctx, "INSERT INTO votes (user_id, resource_id) VALUES ($1, $2)", userId, resourceId)

	return err
}

func (rr ResourceRepo) Downvote(ctx context.Context, userId UserID, resourceId uint) error {
	_, err := rr.db.Exec(ctx, "DELETE FROM votes where user_id=$1 AND resource_id=$2", userId, resourceId)

	return err
}

func (rr ResourceRepo) Search(ctx context.Context, searchQuery string, userId UserID) ([]Resource, error) {
	var resources []Resource

	tsquery := strings.Join(strings.Split(searchQuery, " "), " | ")

	rows, err := rr.db.Query(
		ctx,
		`SELECT
      resource_search.id,
      link,
      name,
      rank,
	    CASE WHEN votes.id IS NULL THEN 0 ELSE 1 END uservoted,
      tags
    FROM
      resource_search
      LEFT JOIN votes on votes.resource_id = resource_search.id AND votes.user_id = $1
    WHERE
       resource_search.doc @@ to_tsquery($2)
    ORDER BY
      rank DESC,
      inserted_at;`,
		userId,
		tsquery,
	)
	defer rows.Close()

	if err != nil {
		return resources, err
	}

	for rows.Next() {
		var id int
		var link string
		var name string
		var rank int
		var voted int
		var tags []string
		rows.Scan(&id, &link, &name, &rank, &voted, &tags)

		urlLink, err := url.Parse(link)

		if err != nil {
			return resources, err
		}

		resources = append(
			resources,
			Resource{
				ID:    uint(id),
				Link:  *urlLink,
				Name:  name,
				Rank:  rank,
				Voted: voted == 1,
				Tags:  tags,
			},
		)
	}

	if err != nil {
		log.Fatal(err.Error())
		return resources, err
	}

	return resources, nil
}

func (rr ResourceRepo) withTransaction(ctx context.Context, fn func() error) error {
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
