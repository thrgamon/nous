package repo

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Comment struct {
	ID    uint
	Content  string
	Username  string
	ParentId UserID
}

type CommentRepo struct {
	db      *pgxpool.Pool
}

func NewCommentRepo(db *pgxpool.Pool) *CommentRepo {
	var repo CommentRepo
	repo.db = db
	return &repo
}

func (rr CommentRepo) GetAll(ctx context.Context, resourceId uint) ([]Comment, error) {
	var comments []Comment

	rows, err := rr.db.Query(
		ctx,
    `WITH RECURSIVE subcomments AS (
	SELECT
		id,
		content,
		user_id,
		parent_id
	FROM
		comments
	WHERE
		resource_id = $1
	UNION
		SELECT
			e.id,
			e.content,
			e.user_id,
			e.parent_id
		FROM
			comments e
		INNER JOIN subcomments s ON s.id = e.parent_id
) SELECT
	subcomments.id,
	content,
	users.username,
	COALESCE(parent_id, 0) as parent_id
FROM
	subcomments
	join users on users.id = user_id;`,
		resourceId,
	)
	defer rows.Close()

	if err != nil {
		return comments, err
	}

	for rows.Next() {
		var id int
		var content string
		var username string
		var parentId uint
		rows.Scan(&id, &content, &username, &parentId)

    comments = append(
      comments, 
      Comment{
        ID: uint(id), 
        Content: content,
        Username: username, 
        ParentId: UserID(parentId), 
      },
	  )
  }

	if rows.Err() != nil {
    log.Fatal(rows.Err().Error())
		return comments, err
	}

	return comments, nil
}
