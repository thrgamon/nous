package repo

import (
	"context"
	"log"
	"sort"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Comment struct {
	ID         uint
	Content    string
	Username   string
	ParentId   uint
	ResourceId ResourceID
	Children   []Comment
}

type CommentRepo struct {
	db *pgxpool.Pool
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
        parent_id desc, inserted_at desc;`,
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
			ID:         uint(id),
			Content:    content,
			Username:   username,
			ParentId:   parentId,
			ResourceId: ResourceID(resourceId),
		}

		// Check to see if there are comments already under this parent root
		_, parentIdPresent := commentTree[parentId]

		// If the parent id is present and it is the root id
		if parentIdPresent && parentId == 0 {
			// If there are child comments for this comment, grab them and assign them to this comment
			children, childrenPresent := commentTree[id]
			if childrenPresent {
				newComment.Children = children
			}
			// Start an array of child comments, including this comment, under the parent comment
			commentTree[parentId] = append(commentTree[parentId], newComment)
			// Delete the child comments key
			delete(commentTree, id)
			// If there are comments grouped under the parent comment then start
			// an array of sibling comments
		} else if parentIdPresent {

			children, childrenPresent := commentTree[id]
			// If there are child comments for this comment, grab them and assign them to this comment
			if childrenPresent {
				newComment.Children = children
			}

			commentTree[parentId] = append(commentTree[parentId], newComment)
			// Delete the child comments key
			delete(commentTree, id)
			// If there are no comments grouped under the parent comment, then check to
			// see if there are child comments for this comment
		} else {
			children, childrenPresent := commentTree[id]
			// If there are child comments for this comment, grab them and assign them to this comment
			if childrenPresent {
				newComment.Children = children
			}
			// Start an array of child comments, including this comment, under the parent comment
			commentTree[parentId] = []Comment{newComment}
			// Delete the child comments key
			delete(commentTree, id)
		}
	}

	root := commentTree[0]
	sort.Slice(root, func(i, j int) bool {
		return root[i].ID < root[j].ID
	})

	if rows.Err() != nil {
		log.Fatal(rows.Err().Error())
		return commentTree, err
	}

	return commentTree, nil
}

func (rr CommentRepo) Add(ctx context.Context, userId UserID, resourceId uint, content string) error {
	_, err := rr.db.Exec(ctx, "INSERT INTO comments (user_id, resource_id, content) VALUES ($1, $2, $3)", userId, resourceId, content)

	return err
}

func (rr CommentRepo) AddChild(ctx context.Context, userId UserID, resourceId uint, parentId uint, content string) error {
	_, err := rr.db.Exec(ctx, "INSERT INTO comments (user_id, resource_id, parent_id, content) VALUES ($1, $2, $3, $4)", userId, resourceId, parentId, content)

	return err
}
