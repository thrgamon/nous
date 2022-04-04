package repo

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Resource struct {
	ID    uint
	Link  string
	Name  string
	Rank  int
	Voted bool
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
      resources.id,
      link,
      name,
      count(votes.user_id) AS rank,
      COUNT(votes.user_id) FILTER (WHERE votes.user_id = $1) AS userVoted
    FROM
      resources
      LEFT JOIN votes ON votes.resource_id = resources.id
    GROUP BY
      resources.id
    ORDER BY
      rank DESC,
      resources.inserted_at;`,
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
		rows.Scan(&id, &link, &name, &rank, &voted)
		resources = append(resources, Resource{ID: uint(id), Link: link, Name: name, Rank: rank, Voted: voted == 1})
	}

	if err != nil {
		return resources, err
	}

	return resources, nil
}

func (rr ResourceRepo) Add(ctx context.Context, link string, name string) error {
	_, err := rr.db.Exec(ctx, "INSERT INTO resources (name, link, rank) VALUES ($1, $2, $3)", name, link, 0)

	return err
}

func (rr ResourceRepo) Upvote(ctx context.Context, userId UserID, resourceId uint) error {
	_, err := rr.db.Exec(ctx, "INSERT INTO votes (user_id, resource_id) VALUES ($1, $2)", userId, resourceId)

	return err
}

func (rr ResourceRepo) Downvote(ctx context.Context, userId UserID, resourceId uint) error {
	_, err := rr.db.Exec(ctx, "DELETE FROM votes where user_id=$1 AND resource_id=$2", userId, resourceId)

	return err
}
