package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Database *pgxpool.Pool

func Init() {
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	Database = conn
}
