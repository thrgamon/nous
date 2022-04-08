package repo

import (
	"context"
	"log"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Resource struct {
	ID    uint
	Link  string
	Name  string
	Rank  int
	Voted bool
  Tags []string
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

func (rr ResourceRepo) Get(id uint) (error, Resource) {
	var link string
	var name string
	var rank int
	err := rr.db.QueryRow(context.TODO(), "select link, name, rank from resources where resources.id = $1", id).Scan(&link, &name, &rank)

	if err != nil {
		return err, Resource{}
	}

	resource := Resource{ID: id, Link: link, Name: name, Rank: rank}

	return nil, resource
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
	tags,
	CASE WHEN votes.id IS NULL THEN 0 ELSE 1 END uservoted
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
    resources = append(
      resources, 
      Resource{
        ID: uint(id), 
        Link: link, 
        Name: name, 
        Rank: rank, 
        Voted: voted == 1,
        Tags: tags,
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
  error := rr.withTransaction(ctx, func() error{
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
      tags,
	    CASE WHEN votes.id IS NULL THEN 0 ELSE 1 END uservoted
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
    resources = append(
      resources, 
      Resource{
        ID: uint(id), 
        Link: link, 
        Name: name, 
        Rank: rank, 
        Voted: voted == 1,
        Tags: tags,
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
