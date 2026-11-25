#!/usr/bin/env bash
# Uso: KEY=sua_chave ./scripts/smoke.sh
HOST=${HOST:-http://127.0.0.1:8080}
P='X-Player-ID: 1'

check() { # descrição, esperado, args do curl...
  desc=$1; want=$2; shift 2
  got=$(curl -s -o /dev/null -w '%{http_code}' "$@")
  if [ "$got" = "$want" ]; then echo "OK   $want  $desc"
  else echo "FALHA esperado $want, veio $got  $desc"; fi
}

A="Authorization: Bearer $KEY"
J='Content-Type: application/json'

check "health público"        200 $HOST/health
check "sem chave"             401 $HOST/items -H "$P"
check "chave inválida"        401 $HOST/items -H "$P" -H 'Authorization: Bearer errada'
check "sem X-Player-ID"       400 $HOST/items -H "$A"
check "listar"                200 $HOST/items -H "$P" -H "$A"
check "criar"                 201 -X POST $HOST/items -H "$P" -H "$A" -H "$J" -d '{"kind":"health_potion","quantity":5}'
check "quantity 0"            400 -X POST $HOST/items -H "$P" -H "$A" -H "$J" -d '{"kind":"health_potion","quantity":0}'
check "kind desconhecido"     400 -X POST $HOST/items -H "$P" -H "$A" -H "$J" -d '{"kind":"nao_existe","quantity":1}'
check "JSON inválido"         400 -X POST $HOST/items -H "$P" -H "$A" -H "$J" -d '{"kind":'
check "id no corpo"           400 -X POST $HOST/items -H "$P" -H "$A" -H "$J" -d '{"id":5,"kind":"health_potion","quantity":1}'
check "buscar inexistente"    404 $HOST/items/99999 -H "$P" -H "$A"
check "id não numérico"       400 $HOST/items/abc -H "$P" -H "$A"
check "buscar sem chave"      401 $HOST/items/1 -H "$P"
check "atualizar inexistente" 404 -X PUT $HOST/items/99999 -H "$P" -H "$A" -H "$J" -d '{"quantity":5}'
check "atualizar sem chave"   401 -X PUT $HOST/items/1 -H "$P" -H "$J" -d '{"quantity":5}'
check "excluir inexistente"   404 -X DELETE $HOST/items/99999 -H "$P" -H "$A"
check "excluir sem chave"     401 -X DELETE $HOST/items/1 -H "$P"
check "método não permitido"  405 -X PATCH $HOST/items/1 -H "$P" -H "$A"
