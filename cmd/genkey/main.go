package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gabriel-Caldeira/back-end_restful/internal/httpapi"
	"github.com/Gabriel-Caldeira/back-end_restful/internal/repo"
)

func main() {
	name := flag.String("name", "", "nome da chave (obrigatório)")
	readOnly := flag.Bool("readonly", false, "chave somente leitura")
	flag.Parse()
	if *name == "" {
		log.Fatal("use -name para identificar a chave")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL não definida")
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("erro ao gerar chave: %v", err)
	}
	key := hex.EncodeToString(buf)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("erro ao criar pool: %v", err)
	}
	defer pool.Close()

	keys := repo.NewKeyDAO(pool)
	if err := keys.Create(ctx, *name, httpapi.HashKey(key), *readOnly); err != nil {
		log.Fatalf("erro ao gravar chave: %v", err)
	}

	fmt.Println("Chave criada. Guarde-a agora, ela não será mostrada de novo:")
	fmt.Println(key)
}
