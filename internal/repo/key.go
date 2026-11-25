package repo

import (
	"context"
	"time"
)

type APIKey struct {
	ID        int64
	Name      string
	ReadOnly  bool
	RevokedAt *time.Time
}

type KeyRepo interface {
	FindByHash(ctx context.Context, hash string) (APIKey, error)
	Create(ctx context.Context, name, hash string, readOnly bool) error
}
