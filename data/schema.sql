-- API de inventário do jogo — estrutura do banco de dados.
-- Execute como proprietário do cluster (postgres), NÃO como inventory_app.

BEGIN;

-- ============================================================
-- Catálogo: o que o jogo sabe sobre cada tipo de item.
-- Dados de referência, compartilhados por todos os jogadores.
-- max_stack = 1 significa que o item não é empilhável.
-- ============================================================
CREATE TABLE item_kinds (
    kind       TEXT    PRIMARY KEY,
    name       TEXT    NOT NULL,
    max_stack  INTEGER NOT NULL CHECK (max_stack >= 1)
);

-- ============================================================
-- Inventário: o que um jogador possui de fato.
-- ============================================================
CREATE TABLE items (
    id          BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    player_id   BIGINT      NOT NULL,
    kind        TEXT        NOT NULL REFERENCES item_kinds(kind),
    quantity    INTEGER     NOT NULL CHECK (quantity > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Snapshot do catálogo no momento da aquisição. Ele reside na linha,
    -- uma vez que uma coluna gerada não pode fazer referência a outra tabela.
    max_stack   INTEGER     NOT NULL CHECK (max_stack >= 1),

    -- Identidade de empilhamento. NULL quando max_stack = 1, de modo que várias espadas
    -- idênticas possam coexistir (um índice UNIQUE ignora valores NULL).
    stack_key   TEXT GENERATED ALWAYS AS (
        CASE WHEN max_stack > 1
             THEN player_id::text || ':' || kind
             ELSE NULL END
    ) STORED
);

CREATE UNIQUE INDEX items_stack_identity ON items(stack_key);
CREATE INDEX        items_player_idx     ON items(player_id);

-- ============================================================
-- Chaves de API: apenas o hash é armazenado, nunca a própria chave.
-- ============================================================
CREATE TABLE api_keys (
    id          BIGINT      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT        NOT NULL,
    key_hash    TEXT        NOT NULL UNIQUE,
    read_only   BOOLEAN     NOT NULL DEFAULT FALSE,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Anti-replay: cada *nonce* de voucher pode ser utilizado apenas uma vez.
-- ============================================================
CREATE TABLE used_nonces (
    nonce       TEXT        PRIMARY KEY,
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;
