package repo

import (
	"context"
	"errors"
)

// ErrNotFound: o handler traduz isso em 404, sem conhecer pgx.
var ErrNotFound = errors.New("não encontrado")

type Kind struct {
	Kind     string
	Name     string
	MaxStack int
}

type KindRepo interface {
	Get(ctx context.Context, kind string) (Kind, error)
}
