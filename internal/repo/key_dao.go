package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type keyDAO struct {
	pool *pgxpool.Pool
}

func NewKeyDAO(pool *pgxpool.Pool) KeyRepo {
	return &keyDAO{pool: pool}
}

func (d *keyDAO) FindByHash(ctx context.Context, hash string) (APIKey, error) {
	var k APIKey
	err := d.pool.QueryRow(ctx,
		"SELECT id, name, read_only, revoked_at FROM api_keys WHERE key_hash = $1",
		hash,
	).Scan(&k.ID, &k.Name, &k.ReadOnly, &k.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return APIKey{}, ErrNotFound
	}
	return k, err
}

func (d *keyDAO) Create(ctx context.Context, name, hash string, readOnly bool) error {
	_, err := d.pool.Exec(ctx,
		"INSERT INTO api_keys (name, key_hash, read_only) VALUES ($1, $2, $3)",
		name, hash, readOnly)
	return err
}
