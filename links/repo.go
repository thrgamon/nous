package links

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thrgamon/nous/database"
	"github.com/thrgamon/nous/logger"
)

type LinkID string

type ArchiveStatus int

const (
	Unsubmitted ArchiveStatus = iota + 1
	Pending
	Error
	Success
)

type Link struct {
	LinkID           LinkID `json:"id"`
	Url              string `json:"url"`
	Title            string `json:"title"`
	ArchiveStatus    int    `json:"archive_status"`
	ArchiveJobID     string `json:"archive_job_id"`
	ArchiveException string `json:"archive_exception"`
}

type LinkRepo struct {
	db     *pgxpool.Pool
	logger *log.Logger
}

func NewLinkRepo() *LinkRepo {
	db := database.Database
	logger := logger.Logger
	return &LinkRepo{db: db, logger: logger}
}

func (lr *LinkRepo) Exists(ctx context.Context, url string) (bool, error) {
	var exists bool
	err := lr.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM links WHERE url_hash=md5($1))`, url).Scan(&exists)
	return exists, err
}

func (lr *LinkRepo) AddLink(ctx context.Context, url string) (LinkID, error) {
	var id int
	err := lr.db.QueryRow(ctx, "INSERT INTO links (url) VALUES ($1) RETURNING id", url).Scan(&id)
	return LinkID(fmt.Sprint(id)), err
}

func (lr *LinkRepo) EditLinkTitle(ctx context.Context, id LinkID, title string) error {
	_, err := lr.db.Exec(ctx, "UPDATE links SET title=$1 WHERE links.id = $2", title, id)
	return err
}

func (lr *LinkRepo) EditLinkURL(ctx context.Context, id LinkID, url string) error {
	_, err := lr.db.Exec(ctx, "UPDATE links SET url=$1 WHERE links.id = $2", url, id)
	return err
}

func (lr *LinkRepo) EditArchiveStatus(ctx context.Context, id LinkID, status ArchiveStatus, jobID string, exception string) error {
	_, err := lr.db.Exec(ctx, "UPDATE links SET archive_status=$1, archive_exception=$2, archive_job_id=$3 WHERE links.id = $4", status, exception, jobID, id)
	return err
}
