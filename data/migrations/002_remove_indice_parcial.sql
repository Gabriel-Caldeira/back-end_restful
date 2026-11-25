BEGIN;
DROP INDEX items_one_partial_stack;
CREATE INDEX items_player_kind_idx ON items (player_id, kind);
COMMIT;
