package storage

import (
	"context"
	"io"
)

type Object struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
}

type Storage interface {
	Backend() string
	Put(ctx context.Context, category Category, filename string, body io.Reader, contentType string) error
	Get(ctx context.Context, category Category, filename string) (*Object, error)
	Delete(ctx context.Context, category Category, filename string) error
	Exists(ctx context.Context, category Category, filename string) (bool, error)
	EnsureDirs() error
}
