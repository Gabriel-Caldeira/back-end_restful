package repo

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUnknownKind = errors.New("tipo de item desconhecido")
	ErrInvalid     = errors.New("dados inválidos")
)

type Item struct {
	ID        int64     `json:"id"`
	PlayerID  int64     `json:"player_id"`
	Kind      string    `json:"kind"`
	Quantity  int       `json:"quantity"`
	MaxStack  int       `json:"max_stack"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemRepo interface {
	Get(ctx context.Context, playerID, id int64) (Item, error)
	List(ctx context.Context, playerID int64) ([]Item, error)
	Create(ctx context.Context, playerID int64, kind string, quantity int) ([]Item, error)
	Update(ctx context.Context, playerID, id int64, quantity int) (Item, error)
	Delete(ctx context.Context, playerID, id int64) error
}
