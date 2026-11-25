package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Gabriel-Caldeira/back-end_restful/internal/repo"
)

var publicRoutes = map[string]bool{
	"GET /health": true,
}

func HashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func Auth(keys repo.KeyRepo, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if publicRoutes[r.Method+" "+r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			writeError(w, http.StatusUnauthorized, "chave de API ausente")
			return
		}

		key, err := keys.FindByHash(r.Context(), HashKey(token))
		if errors.Is(err, repo.ErrNotFound) || (err == nil && key.RevokedAt != nil) {
			writeError(w, http.StatusUnauthorized, "chave de API inválida")
			return
		}
		if err != nil {
			log.Printf("auth: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}

		if key.ReadOnly && r.Method != http.MethodGet {
			writeError(w, http.StatusForbidden, "chave somente leitura")
			return
		}

		next.ServeHTTP(w, r)
	})
}
