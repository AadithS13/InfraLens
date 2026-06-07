package handler

import (
	"net/http"

	"github.com/infralens/infralens/internal/core"
)

type CrawlHandler struct {
	svc *core.ProjectService
}

func NewCrawlHandler(svc *core.ProjectService) *CrawlHandler {
	return &CrawlHandler{svc: svc}
}

// List handles GET /api/v1/crawls
func (h *CrawlHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)

	runs, err := h.svc.ListCrawlRuns(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": runs})
}
