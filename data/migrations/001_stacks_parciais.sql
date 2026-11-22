BEGIN;
DROP INDEX items_stack_identity;
ALTER TABLE items DROP COLUMN stack_key;
CREATE UNIQUE INDEX items_one_partial_stack
    ON items (player_id, kind)
    WHERE quantity < max_stack;
COMMIT;
