# Inventário de Jogo: back-end RESTful

API REST em Go (`net/http`) com PostgreSQL para o inventário de itens de um jogo.

## Requisitos

- Go 1.22 ou superior
- PostgreSQL 14 ou superior

## Como executar do zero

1. **Criar o banco e o usuário da aplicação** (como administrador do PostgreSQL):

```sql
CREATE DATABASE game_inventory;
CREATE USER inventory_app WITH PASSWORD 'CHANGE_ME';
```

Troque `CHANGE_ME` por uma senha própria. Use o mesmo valor no passo 4.

2. **Aplicar o schema** (como administrador):

```bash
psql -d game_inventory -f data/schema.sql
```

3. **Conceder permissões mínimas** ao usuário da aplicação (o schema não inclui `GRANT`):

```bash
psql -d game_inventory -c "GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO inventory_app;"
psql -d game_inventory -c "GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO inventory_app;"
```

O usuário da aplicação não recebe `CREATE`, `DROP` nem `TRUNCATE` (menor privilégio).

4. **Definir as variáveis de ambiente** (copie `.env.example` para `.env` e edite):

```bash
cp .env.example .env
set -a; source .env; set +a
```

| Variável | Descrição | Exemplo |
|---|---|---|
| `DATABASE_URL` | conexão com o banco (usuário da aplicação, nunca `postgres`) | `postgres://inventory_app:CHANGE_ME@127.0.0.1:55432/game_inventory?sslmode=disable` |
| `HTTP_ADDR` | endereço do servidor (opcional, padrão `:8080`) | `127.0.0.1:8080` |

Ajuste a porta conforme o seu PostgreSQL (o padrão é `5432`; o exemplo usa `55432`).

5. **Cadastrar um tipo de item** (o catálogo começa vazio):

```bash
psql -d game_inventory -c "INSERT INTO item_kinds (kind, name, max_stack) VALUES ('health_potion', 'Poção de Vida', 99);"
```

6. **Gerar uma chave de API** (exibida uma única vez; só o hash fica no banco):

```bash
go run ./cmd/genkey -name "servidor-dev"
```

Use `-readonly` para criar uma chave somente leitura.

7. **Subir o servidor:**

```bash
go run ./cmd/server
```

8. **Testar:** `GET /health` responde 200 sem chave. As demais rotas exigem os cabeçalhos `Authorization: Bearer <chave>` e `X-Player-ID`.

## Manual da API

### Convenções

- Corpo e respostas em JSON (`Content-Type: application/json`).
- Todas as rotas, **exceto `GET /health`**, exigem:
  - `Authorization: Bearer <chave>`
  - `X-Player-ID: <inteiro positivo>`: jogador dono do inventário.
- Erros têm o formato `{"error": "mensagem"}`.
- O `id` é gerado pelo servidor; enviá-lo no corpo resulta em 400.
- Item de outro jogador é tratado como inexistente (404).

### Modelo de dados

**Item** (`items`)

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | inteiro | gerado pelo servidor |
| `player_id` | inteiro | dono do item (vem do `X-Player-ID`) |
| `kind` | texto | tipo do item, existente em `item_kinds` |
| `quantity` | inteiro | de 1 até `max_stack` |
| `max_stack` | inteiro | limite do stack, copiado do catálogo na aquisição |
| `created_at` | data/hora | criação |
| `updated_at` | data/hora | última alteração |

**Tipo de item** (`item_kinds`, cadastrado pelo administrador): `kind`, `name`, `max_stack`.

### Regra de empilhamento

Ao criar, o servidor completa o stack incompleto mais antigo do jogador e abre novos stacks para o que sobrar, cada um com até `max_stack`. Por isso o `POST` pode devolver **vários** itens. Itens com `max_stack = 1` geram uma linha por unidade.

### Endpoints

| Método e rota | Corpo | Sucesso | Erros |
|---|---|---|---|
| `GET /health` | n/a | 200 (sem corpo) | n/a |
| `GET /items` | n/a | 200, lista de itens | 400, 401 |
| `POST /items` | `{"kind": "health_potion", "quantity": 20}` | 201, lista de itens afetados, cabeçalho `Location: /items/{id}` | 400, 401, 403 |
| `GET /items/{id}` | n/a | 200, item | 400, 401, 404 |
| `PUT /items/{id}` | `{"quantity": 50}` | 200, item atualizado | 400, 401, 403, 404 |
| `DELETE /items/{id}` | n/a | 204 (sem corpo) | 400, 401, 403, 404 |

### Erros possíveis

| Código | Quando |
|---|---|
| 400 | `X-Player-ID` ausente/inválido; JSON malformado ou com campos desconhecidos; `kind` vazio ou desconhecido; `quantity` ≤ 0 ou acima de `max_stack`; `id` não numérico |
| 401 | chave ausente, inválida ou revogada |
| 403 | chave somente leitura tentando `POST`, `PUT` ou `DELETE` |
| 404 | item inexistente ou de outro jogador |
| 405 | método não permitido na rota |
| 500 | falha interna (sem detalhes ao cliente) |

### Exemplo

```bash
curl -i -X POST http://127.0.0.1:8080/items \
  -H "Authorization: Bearer $KEY" -H "X-Player-ID: 1" \
  -d '{"kind":"health_potion","quantity":20}'
```
