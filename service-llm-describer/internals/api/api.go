package api

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/usercommon/llm-describer/internals/repository"
	"github.com/usercommon/llm-describer/internals/worker"
)

type API struct {
	proc *worker.Processor
	db   *sqlx.DB
}

func NewAPI(proc *worker.Processor, db *sqlx.DB) *API {
	return &API{proc: proc, db: db}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /search", a.handleSearch)
	fs := http.FileServer(http.Dir("./"))
	mux.Handle("GET /", fs)
}

func (a *API) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req repository.SearchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if req.Limit <= 0 {
		req.Limit = 5
	}

	// Get embeddings
	vec, err := a.proc.GenerateEmbedding(r.Context(), req.Query)
	if err != nil {
		http.Error(w, "Embedding failed", http.StatusInternalServerError)
		return
	}

	results, err := repository.SearchPages(a.db, vec, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderJSON(w, results)
}

func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
