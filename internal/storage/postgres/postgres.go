package postgres

import (
	"URL-Shortener/internal/storage"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

//TODO: make stop function

func NewStorage(url string) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	db, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(url, alias string) (int64, error) {
	const op = "storage.postgres.SaveUrl"

	query := "INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id"

	var id int64

	err := s.db.QueryRow(context.Background(), query, url, alias).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return 0, storage.ErrURLExists
		}
		return 0, fmt.Errorf("%s: execute statement: %v", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetUrl"

	query := "SELECT url FROM url WHERE alias = $1"

	var resUrl string

	if err := s.db.QueryRow(context.Background(), query, alias).Scan(&resUrl); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrNotFound
		}
		return "", fmt.Errorf("%s: %v", op, err)
	}

	return resUrl, nil
}

func (s *Storage) DeleteURL(alias string) (int64, error) {
	const op = "storage.postgres.DeleteUrl"

	query := "DELETE FROM url WHERE alias = $1 RETURNING id"

	var id int64

	if err := s.db.QueryRow(context.Background(), query, alias).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, storage.ErrNotFound
		}
		return 0, fmt.Errorf("%s: %v", op, err)
	}

	return id, nil
}
