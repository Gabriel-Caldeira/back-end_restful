package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Gabriel-Caldeira/back-end_restful/internal/repo"
)

type ItemHandler struct {
	items repo.ItemRepo
}

func NewItemHandler(items repo.ItemRepo) *ItemHandler {
	return &ItemHandler{items: items}
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	pid, ok := playerID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "cabeçalho X-Player-ID ausente ou inválido")
		return
	}

	items, err := h.items.List(r.Context(), pid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	pid, ok := playerID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "cabeçalho X-Player-ID ausente ou inválido")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	item, err := h.items.Get(r.Context(), pid, id)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "item não encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

type createRequest struct {
	Kind     string `json:"kind"`
	Quantity int    `json:"quantity"`
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	pid, ok := playerID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "cabeçalho X-Player-ID ausente ou inválido")
		return
	}

	var req createRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Kind == "" {
		writeError(w, http.StatusBadRequest, "kind é obrigatório")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity deve ser maior que zero")
		return
	}

	items, err := h.items.Create(r.Context(), pid, req.Kind, req.Quantity)
	if errors.Is(err, repo.ErrUnknownKind) {
		writeError(w, http.StatusBadRequest, "kind desconhecido")
		return
	}
	if errors.Is(err, repo.ErrInvalid) {
		writeError(w, http.StatusBadRequest, "dados inválidos")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/items/%d", items[0].ID))
	writeJSON(w, http.StatusCreated, items)
}

type updateRequest struct {
	Quantity int `json:"quantity"`
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	pid, ok := playerID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "cabeçalho X-Player-ID ausente ou inválido")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req updateRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	if req.Quantity <= 0 {
		writeError(w, http.StatusBadRequest, "quantity deve ser maior que zero")
		return
	}

	item, err := h.items.Update(r.Context(), pid, id, req.Quantity)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "item não encontrado")
		return
	}
	if errors.Is(err, repo.ErrInvalid) {
		writeError(w, http.StatusBadRequest, "quantity acima do limite do item")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	pid, ok := playerID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "cabeçalho X-Player-ID ausente ou inválido")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	err = h.items.Delete(r.Context(), pid, id)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, http.StatusNotFound, "item não encontrado")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
