// Package workflow builds the HTTP routes for the manager/workflow
// sub-domain. Routes are mounted on a chi.Router and follow the same
// conventions as the alert and edge handlers in this project.
//
// The handler delegates all business logic to a WorkflowService interface
// so tests can swap in a fake without constructing the full biz stack.
package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	biz "github.com/ongridio/ongrid/internal/manager/biz/workflow"
	model "github.com/ongridio/ongrid/internal/manager/model/workflow"
)

// WorkflowService is the narrow service contract the handler depends on.
// *biz.Usecase satisfies it by structural typing.
type WorkflowService interface {
	Create(ctx context.Context, wf *model.Workflow) (*model.Workflow, error)
	Get(ctx context.Context, id string) (*model.Workflow, error)
	List(ctx context.Context) ([]model.Workflow, error)
	Update(ctx context.Context, wf *model.Workflow) (*model.Workflow, error)
	Delete(ctx context.Context, id string) error
	Activate(ctx context.Context, id string) (*model.Workflow, error)
	Pause(ctx context.Context, id string) (*model.Workflow, error)
	Archive(ctx context.Context, id string) (*model.Workflow, error)
	Execute(ctx context.Context, workflowID string, params map[string]interface{}) (*model.WorkflowExecution, error)
	GetExecution(ctx context.Context, executionID string) (*model.WorkflowExecution, error)
	ListExecutions(ctx context.Context, workflowID string) ([]model.WorkflowExecution, error)
	CancelExecution(ctx context.Context, executionID string) error
}

// Handler holds the HTTP routes for workflow CRUD and execution.
type Handler struct {
	svc WorkflowService
}

// NewHandler creates a workflow HTTP handler.
func NewHandler(svc WorkflowService) *Handler {
	return &Handler{svc: svc}
}

// Register mounts all workflow routes onto r.
func (h *Handler) Register(r chi.Router) {
	r.Route("/api/workflows", func(r chi.Router) {
		r.Post("/", h.createWorkflow)
		r.Get("/", h.listWorkflows)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getWorkflow)
			r.Put("/", h.updateWorkflow)
			r.Delete("/", h.deleteWorkflow)
			r.Post("/activate", h.activateWorkflow)
			r.Post("/pause", h.pauseWorkflow)
			r.Post("/execute", h.executeWorkflow)
			r.Get("/executions", h.listExecutions)
		})
	})

	r.Route("/api/workflows/executions", func(r chi.Router) {
		r.Get("/{id}", h.getExecution)
		r.Post("/{id}/cancel", h.cancelExecution)
	})
}

// ---------------------------------------------------------------------------
// Request / response DTOs
// ---------------------------------------------------------------------------

type createWorkflowReq struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Steps       []model.WorkflowStep `json:"steps"`
	Trigger     model.WorkflowTrigger `json:"trigger"`
	CreatedBy   string               `json:"created_by"`
}

type executeReq struct {
	Params map[string]interface{} `json:"params"`
}

type listResp[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (h *Handler) createWorkflow(w http.ResponseWriter, r *http.Request) {
	var req createWorkflowReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, errInvalid(err))
		return
	}
	wf := &model.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Steps:       req.Steps,
		Trigger:     req.Trigger,
		CreatedBy:   req.CreatedBy,
	}
	created, err := h.svc.Create(r.Context(), wf)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listWorkflows(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listResp[model.Workflow]{Items: items, Total: len(items)})
}

func (h *Handler) getWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) updateWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var wf model.Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		writeErr(w, errInvalid(err))
		return
	}
	wf.ID = id
	updated, err := h.svc.Update(r.Context(), &wf)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) activateWorkflow(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.svc.Activate)
}

func (h *Handler) pauseWorkflow(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.svc.Pause)
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request, fn func(context.Context, string) (*model.Workflow, error)) {
	id := chi.URLParam(r, "id")
	wf, err := fn(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, wf)
}

func (h *Handler) executeWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req executeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body — default to empty params.
		req.Params = nil
	}
	if req.Params == nil {
		req.Params = map[string]interface{}{}
	}
	ex, err := h.svc.Execute(r.Context(), id, req.Params)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ex)
}

func (h *Handler) listExecutions(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "id")
	items, err := h.svc.ListExecutions(r.Context(), workflowID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, listResp[model.WorkflowExecution]{Items: items, Total: len(items)})
}

func (h *Handler) getExecution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ex, err := h.svc.GetExecution(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (h *Handler) cancelExecution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.CancelExecution(r.Context(), id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, err error) {
	code := errCode(err)
	writeJSON(w, code, errorBody{Error: err.Error(), Code: errSlug(err)})
}

func errCode(err error) int {
	switch {
	case errors.Is(err, biz.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, biz.ErrInvalid):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func errSlug(err error) string {
	switch {
	case errors.Is(err, biz.ErrNotFound):
		return "not-found"
	case errors.Is(err, biz.ErrInvalid):
		return "invalid-argument"
	default:
		return "internal"
	}
}

func errInvalid(inner error) error {
	return fmt.Errorf("%w: %s", biz.ErrInvalid, inner)
}
