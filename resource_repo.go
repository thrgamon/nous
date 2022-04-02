package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Resource struct {
	ID   uint
	Link string
	Name string
	Rank int
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

func (rr ResourceRepo) GetAll(ctx context.Context) ([]Resource, error) {
	var resources []Resource

	rows, err := rr.db.Query(ctx, "SELECT id, link, name, rank FROM resources ORDER BY rank DESC")
	defer rows.Close()

	if err != nil {
		return resources, err
	}

	for rows.Next() {
		var id int
		var link string
		var name string
		var rank int
		rows.Scan(&id, &link, &name, &rank)
		resources = append(resources, Resource{ID: uint(id), Link: link, Name: name, Rank: rank})
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

func (rr ResourceRepo) Upvote(ctx context.Context, id uint) error {
	_, err := rr.db.Exec(ctx, "UPDATE resources SET rank = rank + 1 WHERE id = $1", id)

	return err
}

func (rr ResourceRepo) Downvote(ctx context.Context, id uint) error {
	_, err := rr.db.Exec(ctx, "UPDATE resources SET rank = rank - 1 WHERE id = $1", id)

	return err
}
