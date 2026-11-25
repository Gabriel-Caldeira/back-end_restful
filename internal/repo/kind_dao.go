package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type kindDAO struct {
	pool *pgxpool.Pool
}

func NewKindDAO(pool *pgxpool.Pool) KindRepo {
	return &kindDAO{pool: pool}
}

func (d *kindDAO) Get(ctx context.Context, kind string) (Kind, error) {
	var k Kind
	err := d.pool.QueryRow(ctx,
		"SELECT kind, name, max_stack FROM item_kinds WHERE kind = $1",
		kind,
	).Scan(&k.Kind, &k.Name, &k.MaxStack)

	if errors.Is(err, pgx.ErrNoRows) {
		return Kind{}, ErrNotFound
	}
	if err != nil {
		return Kind{}, err
	}
	return k, nil
}
