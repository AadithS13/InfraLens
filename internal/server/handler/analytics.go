package handler

import (
	"net/http"

	"github.com/infralens/infralens/internal/core"
)

type AnalyticsHandler struct {
	svc *core.ProjectService
}

func NewAnalyticsHandler(svc *core.ProjectService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// StatusDistribution handles GET /api/v1/analytics/status-distribution
func (h *AnalyticsHandler) StatusDistribution(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.StatusDistribution(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// TopBuilders handles GET /api/v1/analytics/top-builders
func (h *AnalyticsHandler) TopBuilders(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 10)
	items, err := h.svc.TopBuilders(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// ByDistrict handles GET /api/v1/analytics/by-district
func (h *AnalyticsHandler) ByDistrict(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	items, err := h.svc.ByDistrict(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}
