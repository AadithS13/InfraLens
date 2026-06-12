package handler

import (
	"net/http"

	"github.com/infralens/infralens/internal/core"
	"github.com/infralens/infralens/internal/nlsearch"
)

// NLSearchHandler handles GET /api/v1/search/nl — V7.1 Rule-Based NL Search.
//
// The query string is parsed by the nlsearch package into a structured filter.
// The response always includes an "interpreted" field showing exactly which
// SQL filters were extracted, so callers can show "Searched for: status=Ongoing, district=Pune".
//
// V7.2 will swap in an LLM-backed parser; this handler and its response shape
// stay exactly the same.
type NLSearchHandler struct {
	svc *core.ProjectService
}

func NewNLSearchHandler(svc *core.ProjectService) *NLSearchHandler {
	return &NLSearchHandler{svc: svc}
}

// Search handles GET /api/v1/search/nl?q=<natural language query>
func (h *NLSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required — e.g. ?q=show+ongoing+projects+in+Pune")
		return
	}

	parsed := nlsearch.Parse(q)

	// Builder intent — returns promoter list, not project list
	if parsed.QueryType == "builders" {
		items, err := h.svc.BuildersWithMinProjects(r.Context(), parsed.MinProjects, 20)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"query":       parsed.Raw,
			"query_type":  parsed.QueryType,
			"interpreted": parsed.Interpreted,
			"data":        items,
		})
		return
	}

	// Project intent — run through the standard search layer
	parsed.Filter.Page = queryInt(r, "page", 1)
	parsed.Filter.Limit = queryInt(r, "limit", 20)

	result, err := h.svc.Search(r.Context(), parsed.Filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query":       parsed.Raw,
		"query_type":  parsed.QueryType,
		"interpreted": parsed.Interpreted,
		"data":        result.Data,
		"meta":        result.Meta,
	})
}
