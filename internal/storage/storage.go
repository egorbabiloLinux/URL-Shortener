package storage

import "errors"

var (
	ErrNotFound  = errors.New("url not found")
	ErrURLExists = errors.New("url exists")
)
