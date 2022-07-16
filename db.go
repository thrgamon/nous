package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

func initDB() *pgxpool.Pool {
	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	return conn
}
