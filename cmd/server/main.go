package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Gabriel-Caldeira/back-end_restful/internal/httpapi"
	"github.com/Gabriel-Caldeira/back-end_restful/internal/repo"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL não definida")
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("erro ao criar pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("banco inacessível: %v", err)
	}

	items := httpapi.NewItemHandler(repo.NewItemDAO(pool))
	keys := repo.NewKeyDAO(pool)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /items", items.List)
	mux.HandleFunc("POST /items", items.Create)
	mux.HandleFunc("GET /items/{id}", items.Get)
	mux.HandleFunc("PUT /items/{id}", items.Update)
	mux.HandleFunc("DELETE /items/{id}", items.Delete)

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.Auth(keys, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("ouvindo em %s", addr)
	log.Fatal(srv.ListenAndServe())
}
