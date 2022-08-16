package contexts

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/thrgamon/nous/database"
	"github.com/thrgamon/nous/logger"
)

type ContextRepo struct {
	db     *pgxpool.Pool
	logger *log.Logger
}

func NewContextRepo() *ContextRepo {
	db := database.Database
	logger := logger.Logger
	return &ContextRepo{db: db, logger: logger}
}

func (rr ContextRepo) GetContexts(ctx context.Context) []string {
	rows, err := rr.db.Query(ctx, `SELECT context from contexts`)

	defer rows.Close()

	if err != nil {
		panic(err)
	}

	return rr.parseData(rows)
}

func (rr ContextRepo) GetActiveContext(ctx context.Context) (context string) {
	err := rr.db.QueryRow(ctx, `SELECT context from contexts where active = true`).Scan(&context)

	if err != nil {
		panic(err)
	}

	return context
}

func (rr ContextRepo) UpdateContext(ctx context.Context, context string) {
	_, err := rr.db.Exec(ctx, `update contexts set active = false where active = true;`)
	if err != nil {
		panic(err)
	}

	_, err = rr.db.Exec(ctx, `update contexts set active = true where context = $1;`, context)
	if err != nil {
		panic(err)
	}
}

func (rr ContextRepo) parseData(rows pgx.Rows) []string {
	var contexts []string

	for rows.Next() {
		var context string
		err := rows.Scan(&context)

		if err != nil {
			panic(err)
		}

		contexts = append(contexts, context)
	}

	return contexts
}
