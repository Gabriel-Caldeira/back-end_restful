package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type itemDAO struct {
	pool *pgxpool.Pool
}

func NewItemDAO(pool *pgxpool.Pool) ItemRepo {
	return &itemDAO{pool: pool}
}

const itemColumns = "id, player_id, kind, quantity, max_stack, created_at, updated_at"

func scanItem(row pgx.Row) (Item, error) {
	var it Item
	err := row.Scan(&it.ID, &it.PlayerID, &it.Kind, &it.Quantity,
		&it.MaxStack, &it.CreatedAt, &it.UpdatedAt)
	return it, err
}

// GET
func (d *itemDAO) Get(ctx context.Context, playerID, id int64) (Item, error) {
	row := d.pool.QueryRow(ctx,
		"SELECT "+itemColumns+" FROM items WHERE id = $1 AND player_id = $2",
		id, playerID)
	it, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	return it, err
}

// LIST
func (d *itemDAO) List(ctx context.Context, playerID int64) ([]Item, error) {
	rows, err := d.pool.Query(ctx,
		"SELECT "+itemColumns+" FROM items WHERE player_id = $1 ORDER BY id",
		playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// DELETE
func (d *itemDAO) Delete(ctx context.Context, playerID, id int64) error {
	tag, err := d.pool.Exec(ctx,
		"DELETE FROM items WHERE id = $1 AND player_id = $2",
		id, playerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UPDATE
func (d *itemDAO) Update(ctx context.Context, playerID, id int64, quantity int) (Item, error) {
	current, err := d.Get(ctx, playerID, id)
	if err != nil {
		return Item{}, err
	}
	if quantity <= 0 || quantity > current.MaxStack {
		return Item{}, ErrInvalid
	}

	row := d.pool.QueryRow(ctx,
		`UPDATE items
		    SET quantity = $1, updated_at = now()
		  WHERE id = $2 AND player_id = $3
		RETURNING `+itemColumns,
		quantity, id, playerID)

	it, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	return it, err
}

// CREATE
func (d *itemDAO) Create(ctx context.Context, playerID int64, kind string, quantity int) ([]Item, error) {
	if quantity <= 0 {
		return nil, ErrInvalid
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Passo 1: limite do tipo no catálogo
	var maxStack int
	err = tx.QueryRow(ctx,
		"SELECT max_stack FROM item_kinds WHERE kind = $1", kind,
	).Scan(&maxStack)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnknownKind
	}
	if err != nil {
		return nil, err
	}

	result := []Item{}
	remaining := quantity

	// Passo 2: stack parcial do jogador, com trava
	row := tx.QueryRow(ctx,
		"SELECT "+itemColumns+` FROM items
		  WHERE player_id = $1 AND kind = $2 AND quantity < max_stack
		  ORDER BY id
		  LIMIT 1
		  FOR UPDATE`,
		playerID, kind)
	partial, err := scanItem(row)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Passo 3: completar o parcial, se existir
	if err == nil {
		add := min(remaining, partial.MaxStack-partial.Quantity)
		row := tx.QueryRow(ctx,
			`UPDATE items
			    SET quantity = quantity + $1, updated_at = now()
			  WHERE id = $2
			RETURNING `+itemColumns,
			add, partial.ID)
		updated, err := scanItem(row)
		if err != nil {
			return nil, err
		}
		result = append(result, updated)
		remaining -= add
	}

	// Passo 4: novos stacks para o que sobrou
	for remaining > 0 {
		chunk := min(remaining, maxStack)
		row := tx.QueryRow(ctx,
			`INSERT INTO items (player_id, kind, quantity, max_stack)
			 VALUES ($1, $2, $3, $4)
			 RETURNING `+itemColumns,
			playerID, kind, chunk, maxStack)
		created, err := scanItem(row)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
		remaining -= chunk
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
