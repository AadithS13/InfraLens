package handler

import (
	"net/http"

	"github.com/infralens/infralens/internal/core"
)

// SearchHandler handles the /api/v1/search routes.
type SearchHandler struct {
	svc *core.ProjectService
}

func NewSearchHandler(svc *core.ProjectService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Suggestions handles GET /api/v1/search/suggestions?q=<term>&limit=10
//
// Returns ranked autocomplete results (project names + promoter names)
// using pg_trgm word_similarity. Requires at least 2 characters to avoid
// returning the entire dataset on a single-char query.
func (h *SearchHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		writeJSON(w, http.StatusOK, map[string]any{"data": []any{}})
		return
	}

	limit := queryInt(r, "limit", 10)

	items, err := h.svc.Suggestions(r.Context(), q, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}
