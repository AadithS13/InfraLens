package handler

import (
	"net/http"

	"github.com/infralens/infralens/internal/core"
	"github.com/infralens/infralens/internal/nlsearch"
)

// NLSearchHandler handles GET /api/v1/search/nl.
//
// V7.1: parser is a RuleParser (zero cost, no external calls).
// V7.2: parser is an LLMParser wrapped in a FallbackParser (Claude Haiku API).
//
// The response shape, SQL, and handler logic are identical in both versions —
// only the nlsearch.Parser implementation changes.
type NLSearchHandler struct {
	svc    *core.ProjectService
	parser nlsearch.Parser
}

// NewNLSearchHandler wires the handler with the given parser.
// main.go selects either RuleParser or LLMParser+FallbackParser based on
// whether ANTHROPIC_API_KEY is set.
func NewNLSearchHandler(svc *core.ProjectService, parser nlsearch.Parser) *NLSearchHandler {
	return &NLSearchHandler{svc: svc, parser: parser}
}

// Search handles GET /api/v1/search/nl?q=<natural language query>
//
// Response shape:
//
//	{
//	  "query":       "show ongoing projects in Pune",
//	  "query_type":  "projects",
//	  "interpreted": {"status": "Ongoing", "district": "Pune"},
//	  "data":        [...],
//	  "meta":        {...}
//	}
func (h *NLSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required — e.g. ?q=show+ongoing+projects+in+Pune")
		return
	}

	parsed, err := h.parser.Parse(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not parse query: "+err.Error())
		return
	}

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
