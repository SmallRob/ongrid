// Package change builds the HTTP routes for the ITSM change management
// sub-domain. The Handler delegates all business logic to the biz/change
// Usecase; it is responsible only for JSON marshalling, input validation,
// and HTTP status mapping.
package change

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	changebiz "github.com/ongridio/ongrid/internal/manager/biz/change"
	changemodel "github.com/ongridio/ongrid/internal/manager/model/change"
	"github.com/ongridio/ongrid/internal/pkg/errs"
)

// Handler exposes /api/changes.
type Handler struct {
	uc changebiz.ChangeUsecase
}

// NewHandler builds the handler around a ChangeUsecase.
func NewHandler(uc changebiz.ChangeUsecase) *Handler {
	return &Handler{uc: uc}
}

// Register attaches the change management routes on r.
//
// Routes:
//
//	POST   /api/changes                       — create a change request
//	GET    /api/changes                       — list change requests
//	GET    /api/changes/{id}                  — get a single change request
//	PUT    /api/changes/{id}                  — update a change request
//	POST   /api/changes/{id}/approve          — approve
//	POST   /api/changes/{id}/reject           — reject
//	POST   /api/changes/{id}/implement        — start implementation
//	POST   /api/changes/{id}/complete         — mark as completed
//	POST   /api/changes/{id}/cancel           — cancel
//	GET    /api/changes/{id}/audit            — get audit trail
func (h *Handler) Register(r chi.Router) {
	r.Post("/api/changes", h.create)
	r.Get("/api/changes", h.list)
	r.Get("/api/changes/{id}", h.get)
	r.Put("/api/changes/{id}", h.update)
	r.Post("/api/changes/{id}/approve", h.approve)
	r.Post("/api/changes/{id}/reject", h.reject)
	r.Post("/api/changes/{id}/implement", h.implement)
	r.Post("/api/changes/{id}/complete", h.complete)
	r.Post("/api/changes/{id}/cancel", h.cancel)
	r.Get("/api/changes/{id}/audit", h.getAudit)
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type createReq struct {
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Priority       string     `json:"priority"`
	RiskLevel      string     `json:"risk_level"`
	Type           string     `json:"type"`
	AssigneeID     string     `json:"assignee_id,omitempty"`
	CreatedBy      string     `json:"created_by"`
	Implementation string     `json:"implementation,omitempty"`
	RollbackPlan   string     `json:"rollback_plan,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
}

type updateReq struct {
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Priority       *string    `json:"priority,omitempty"`
	RiskLevel      *string    `json:"risk_level,omitempty"`
	AssigneeID     *string    `json:"assignee_id,omitempty"`
	Implementation *string    `json:"implementation,omitempty"`
	RollbackPlan   *string    `json:"rollback_plan,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
}

type actionReq struct {
	ApproverID string `json:"approver_id,omitempty"`
	AssigneeID string `json:"assignee_id,omitempty"`
	Actor      string `json:"actor,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type changeItem struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	RiskLevel      string     `json:"risk_level"`
	Type           string     `json:"type"`
	ApproverID     string     `json:"approver_id,omitempty"`
	AssigneeID     string     `json:"assignee_id,omitempty"`
	CreatedBy      string     `json:"created_by"`
	Implementation string     `json:"implementation,omitempty"`
	RollbackPlan   string     `json:"rollback_plan,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type listResp struct {
	Items []changeItem `json:"items"`
	Total int64        `json:"total"`
}

type auditItem struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	Actor       string    `json:"actor"`
	OldStatus   string    `json:"old_status,omitempty"`
	NewStatus   string    `json:"new_status,omitempty"`
	Description string    `json:"description,omitempty"`
}

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in createReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}

	cr, err := h.uc.Create(r.Context(), changebiz.CreateInput{
		Title:          in.Title,
		Description:    in.Description,
		Priority:       in.Priority,
		RiskLevel:      in.RiskLevel,
		Type:           in.Type,
		AssigneeID:     in.AssigneeID,
		CreatedBy:      in.CreatedBy,
		Implementation: in.Implementation,
		RollbackPlan:   in.RollbackPlan,
		ScheduledAt:    in.ScheduledAt,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, crToItem(cr))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cr, err := h.uc.Get(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := changebiz.ChangeFilter{
		Status:   q.Get("status"),
		Priority: q.Get("priority"),
		Type:     q.Get("type"),
		Assignee: q.Get("assignee"),
		Creator:  q.Get("creator"),
	}
	items, total, err := h.uc.List(r.Context(), f)
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]changeItem, 0, len(items))
	for _, cr := range items {
		out = append(out, crToItem(cr))
	}
	writeJSON(w, http.StatusOK, listResp{Items: out, Total: total})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in updateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}

	cr, err := h.uc.Update(r.Context(), id, changebiz.UpdateInput{
		Title:          in.Title,
		Description:    in.Description,
		Priority:       in.Priority,
		RiskLevel:      in.RiskLevel,
		AssigneeID:     in.AssigneeID,
		Implementation: in.Implementation,
		RollbackPlan:   in.RollbackPlan,
		ScheduledAt:    in.ScheduledAt,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) approve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in actionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}
	cr, err := h.uc.Approve(r.Context(), id, in.ApproverID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) reject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in actionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}
	cr, err := h.uc.Reject(r.Context(), id, in.ApproverID, in.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) implement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in actionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}
	cr, err := h.uc.Implement(r.Context(), id, in.AssigneeID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in actionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}
	cr, err := h.uc.Complete(r.Context(), id, in.Actor)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in actionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, errors.Join(errs.ErrInvalid, err))
		return
	}
	cr, err := h.uc.Cancel(r.Context(), id, in.Actor, in.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, crToItem(cr))
}

func (h *Handler) getAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entries, err := h.uc.GetAuditLog(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]auditItem, 0, len(entries))
	for _, e := range entries {
		out = append(out, auditItem{
			Timestamp:   e.Timestamp,
			Action:      e.Action,
			Actor:       e.Actor,
			OldStatus:   e.OldStatus,
			NewStatus:   e.NewStatus,
			Description: e.Description,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func crToItem(cr *changemodel.ChangeRequest) changeItem {
	return changeItem{
		ID:             cr.ID,
		Title:          cr.Title,
		Description:    cr.Description,
		Status:         cr.Status,
		Priority:       cr.Priority,
		RiskLevel:      cr.RiskLevel,
		Type:           cr.Type,
		ApproverID:     cr.ApproverID,
		AssigneeID:     cr.AssigneeID,
		CreatedBy:      cr.CreatedBy,
		Implementation: cr.Implementation,
		RollbackPlan:   cr.RollbackPlan,
		ScheduledAt:    cr.ScheduledAt,
		CreatedAt:      cr.CreatedAt,
		UpdatedAt:      cr.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, err error) {
	status := errs.HTTPStatus(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{
		Error: err.Error(),
		Code:  errCode(err),
	})
}

func errCode(err error) string {
	switch {
	case errors.Is(err, errs.ErrNotFound):
		return "not-found"
	case errors.Is(err, errs.ErrUnauthorized):
		return "unauthorized"
	case errors.Is(err, errs.ErrForbidden):
		return "forbidden"
	case errors.Is(err, errs.ErrInvalid):
		return "invalid"
	case errors.Is(err, errs.ErrConflict):
		return "conflict"
	default:
		return "internal"
	}
}


