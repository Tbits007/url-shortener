package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)
	

type Storage struct {
	db *sql.DB
}

func New(connStr string) (*Storage, error) {
    const op = "storage.postgres.New"

    db, err := sql.Open("postgres", connStr) 
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    query := `
    CREATE TABLE IF NOT EXISTS url(
        id INTEGER PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL);
    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
    `

    _, err = db.Exec(query)
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }

    return &Storage{db: db}, nil
}