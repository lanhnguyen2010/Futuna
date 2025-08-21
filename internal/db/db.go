package db

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Connect opens a PostgreSQL database connection using the supplied URL.
func Connect(url string) *sqlx.DB {
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	return db
}
