package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Tbits007/url-shortener/internal/storage"
	"github.com/lib/pq"
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

func (s *Storage) SaveURL(urlToSave string, alias string) error {
    const op = "storage.postgres.SaveURL"

    query := `INSERT INTO url(url,alias) values($1,$2)`
    
    _, err := s.db.Exec(query, urlToSave, alias)
    if err != nil {
        if pgErr, ok := err.(*pq.Error); ok {
            if pgErr.Code == "23505" { // 23505 - это код ошибки unique_violation в PostgreSQL
                return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
            }
        }
        return fmt.Errorf("%s: execute query: %w", op, err)
    }

    return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
    const op = "storage.postgres.GetURL"

    query := `SELECT url FROM url WHERE alias = $1`

    row := s.db.QueryRow(query, alias)

    var res string 
    err := row.Scan(&res)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", fmt.Errorf("%s: url not found: %w", op, storage.ErrURLNotFound)
        }
        return "", fmt.Errorf("%s: execute query:%w", op, err)
    } 

    return res, nil 
}

func (s *Storage) DeleteURL(alias string) error {
    const op = "storage.postgres.DeleteURL"
    
    query := `DELETE FROM url WHERE alias = $1`

    res, err := s.db.Exec(query)
    if err != nil {
        return fmt.Errorf("%s: execute query: %w", op, storage.ErrURLNotFound) 
    }

    count, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
    }
    if count == 0 {
        return fmt.Errorf("%s: url not found:%w", op, storage.ErrURLNotFound)
    }   

    return nil 
}
