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
	ParentId uint
  ResourceId ResourceID
  Children []Comment
}

type CommentRepo struct {
	db      *pgxpool.Pool
}

func NewCommentRepo(db *pgxpool.Pool) *CommentRepo {
	var repo CommentRepo
	repo.db = db
	return &repo
}

func (rr CommentRepo) GetAll(ctx context.Context, resourceId uint) (map[uint][]Comment, error) {
  commentTree := make(map[uint][]Comment)

	rows, err := rr.db.Query(
		ctx,
    `WITH RECURSIVE subcomments AS (
      SELECT
        id,
        content,
        user_id,
        parent_id,
        resource_id
      FROM
        comments
      WHERE
        resource_id = $1
      UNION
        SELECT
          e.id,
          e.content,
          e.user_id,
          e.parent_id,
          e.resource_id
        FROM
          comments e
        INNER JOIN subcomments s ON s.id = e.parent_id
    ) SELECT
      subcomments.id,
      content,
      users.username,
      COALESCE(parent_id, 0) as parent_id,
      resource_id
    FROM
      subcomments
      join users on users.id = user_id
    ORDER BY
        parent_id desc;`,
		resourceId,
	)
	defer rows.Close()

	if err != nil {
		return commentTree, err
	}

	for rows.Next() {
		var id uint
		var content string
		var username string
		var parentId uint
		var resourceId uint
		rows.Scan(&id, &content, &username, &parentId, &resourceId)

    newComment := Comment{
        ID: uint(id), 
        Content: content,
        Username: username, 
        ParentId: parentId, 
        ResourceId: ResourceID(resourceId),
      }

    _, parentIdPresent := commentTree[parentId]

    if parentIdPresent{
      commentTree[parentId] = append(commentTree[parentId], newComment)
    } else {
      children, childrenPresent := commentTree[id]
      if childrenPresent {
        newComment.Children = children
      }
      commentTree[parentId] = append(commentTree[parentId], newComment)
      delete(commentTree, id)
    }
  }

	if rows.Err() != nil {
    log.Fatal(rows.Err().Error())
		return commentTree, err
	}

	return commentTree, nil
}

func (rr CommentRepo) Add(ctx context.Context, userId UserID, resourceId uint, content string) error {
  _, err := rr.db.Exec(ctx, "INSERT INTO comments (user_id, resource_id, content) VALUES ($1, $2, $4)", userId, resourceId, content)

  return err
}

func (rr CommentRepo) AddChild(ctx context.Context, userId UserID, resourceId uint, parentId uint, content string) error {
  _, err := rr.db.Exec(ctx, "INSERT INTO comments (user_id, resource_id, parent_id, content) VALUES ($1, $2, $3, $4)", userId, resourceId, parentId, content)

  return err
}

