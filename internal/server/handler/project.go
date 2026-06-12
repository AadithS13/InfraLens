package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/infralens/infralens/internal/core"
)

type ProjectHandler struct {
	svc *core.ProjectService
}

func NewProjectHandler(svc *core.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List handles GET /api/v1/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	f := core.SearchFilter{
		Q:        queryParam(r, "q"),
		City:     queryParam(r, "city"),
		District: queryParam(r, "district"),
		State:    queryParam(r, "state"),
		Promoter: queryParam(r, "promoter"),
		Status:   queryParam(r, "status"),
		Type:     queryParam(r, "type"),
		Page:     queryInt(r, "page", 1),
		Limit:    queryInt(r, "limit", 20),
	}

	result, err := h.svc.Search(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Get handles GET /api/v1/projects/{id}
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	project, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if project == nil {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": project})
}

// Changes handles GET /api/v1/projects/{id}/changes
func (h *ProjectHandler) Changes(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	changes, err := h.svc.GetChanges(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": changes})
}

// --- helpers ---

func queryParam(r *http.Request, key string) *string {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	return &v
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
