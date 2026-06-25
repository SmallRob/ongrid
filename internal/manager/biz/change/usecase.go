package change

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	model "github.com/ongridio/ongrid/internal/manager/model/change"
	"github.com/ongridio/ongrid/internal/pkg/errs"
)

// ChangeUsecase is the facade contract the server layer programs against.
// Implementations must be safe for concurrent use.
type ChangeUsecase interface {
	Create(ctx context.Context, in CreateInput) (*model.ChangeRequest, error)
	Get(ctx context.Context, id string) (*model.ChangeRequest, error)
	List(ctx context.Context, f ChangeFilter) ([]*model.ChangeRequest, int64, error)
	Update(ctx context.Context, id string, in UpdateInput) (*model.ChangeRequest, error)
	Approve(ctx context.Context, id, approverID string) (*model.ChangeRequest, error)
	Reject(ctx context.Context, id, approverID, reason string) (*model.ChangeRequest, error)
	Implement(ctx context.Context, id, assigneeID string) (*model.ChangeRequest, error)
	Complete(ctx context.Context, id, actor string) (*model.ChangeRequest, error)
	Cancel(ctx context.Context, id, actor, reason string) (*model.ChangeRequest, error)
	GetAuditLog(ctx context.Context, id string) ([]*model.AuditEntry, error)
}

// CreateInput carries the fields for a new change request.
type CreateInput struct {
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

// UpdateInput carries the mutable fields for an update. Only non-nil
// pointer fields are applied; the rest are left at their current values.
type UpdateInput struct {
	Title          *string    `json:"title,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Priority       *string    `json:"priority,omitempty"`
	RiskLevel      *string    `json:"risk_level,omitempty"`
	AssigneeID     *string    `json:"assignee_id,omitempty"`
	Implementation *string    `json:"implementation,omitempty"`
	RollbackPlan   *string    `json:"rollback_plan,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
}

// ---------------------------------------------------------------------------
// In-memory Repository implementations
// ---------------------------------------------------------------------------

// MemChangeRepo is a thread-safe in-memory implementation of Repo.
// Suitable for development, unit tests and single-instance deployments.
type MemChangeRepo struct {
	mu      sync.RWMutex
	entries map[string]*model.ChangeRequest
}

// MemAuditRepo is a thread-safe in-memory implementation of AuditRepo.
type MemAuditRepo struct {
	mu      sync.RWMutex
	entries map[string][]*model.AuditEntry
}

// NewMemChangeRepo constructs an empty in-memory change repo.
func NewMemChangeRepo() *MemChangeRepo {
	return &MemChangeRepo{entries: make(map[string]*model.ChangeRequest)}
}

// NewMemAuditRepo constructs an empty in-memory audit repo.
func NewMemAuditRepo() *MemAuditRepo {
	return &MemAuditRepo{entries: make(map[string][]*model.AuditEntry)}
}

// --- MemChangeRepo (Repo) ---

func (r *MemChangeRepo) Create(_ context.Context, cr *model.ChangeRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.entries[cr.ID]; exists {
		return fmt.Errorf("%w: change %s already exists", errs.ErrConflict, cr.ID)
	}
	r.entries[cr.ID] = cr
	return nil
}

func (r *MemChangeRepo) Get(_ context.Context, id string) (*model.ChangeRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cr, ok := r.entries[id]
	if !ok {
		return nil, fmt.Errorf("%w: change %s", errs.ErrNotFound, id)
	}
	// Return a copy so callers cannot mutate the stored value.
	out := *cr
	return &out, nil
}

func (r *MemChangeRepo) List(_ context.Context, f ChangeFilter) ([]*model.ChangeRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []*model.ChangeRequest
	for _, cr := range r.entries {
		if f.Status != "" && cr.Status != f.Status {
			continue
		}
		if f.Priority != "" && cr.Priority != f.Priority {
			continue
		}
		if f.Type != "" && cr.Type != f.Type {
			continue
		}
		if f.Assignee != "" && cr.AssigneeID != f.Assignee {
			continue
		}
		if f.Creator != "" && cr.CreatedBy != f.Creator {
			continue
		}
		cp := *cr
		out = append(out, &cp)
	}

	// Sort by CreatedAt descending (newest first).
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	// Apply pagination.
	if f.Offset > 0 && f.Offset < len(out) {
		out = out[f.Offset:]
	} else if f.Offset >= len(out) {
		return nil, nil
	}
	if f.Limit > 0 && f.Limit < len(out) {
		out = out[:f.Limit]
	}

	return out, nil
}

func (r *MemChangeRepo) Update(_ context.Context, cr *model.ChangeRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[cr.ID]; !ok {
		return fmt.Errorf("%w: change %s", errs.ErrNotFound, cr.ID)
	}
	r.entries[cr.ID] = cr
	return nil
}

func (r *MemChangeRepo) Count(_ context.Context) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.entries)), nil
}

// --- MemAuditRepo (AuditRepo) ---

func (r *MemAuditRepo) Append(_ context.Context, changeID string, entry *model.AuditEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[changeID] = append(r.entries[changeID], entry)
	return nil
}

func (r *MemAuditRepo) List(_ context.Context, changeID string) ([]*model.AuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := r.entries[changeID]
	out := make([]*model.AuditEntry, len(entries))
	copy(out, entries)
	return out, nil
}

// ---------------------------------------------------------------------------
// Usecase
// ---------------------------------------------------------------------------

// Usecase is the biz-layer facade for change management.
type Usecase struct {
	repo  Repo
	audit AuditRepo
	clock Clock
	log   *slog.Logger
}

// Clock abstracts time.Now for testability.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// NewUsecase builds the change management Usecase. Both repo and audit
// are required. log may be nil (defaults to slog.Default).
func NewUsecase(repo Repo, audit AuditRepo, log *slog.Logger) *Usecase {
	if log == nil {
		log = slog.Default()
	}
	return &Usecase{repo: repo, audit: audit, clock: realClock{}, log: log}
}

// NewUsecaseWithClock is the test-friendly constructor.
func NewUsecaseWithClock(repo Repo, audit AuditRepo, log *slog.Logger, clock Clock) *Usecase {
	if log == nil {
		log = slog.Default()
	}
	return &Usecase{repo: repo, audit: audit, clock: clock, log: log}
}

// Create validates and persists a new ChangeRequest. Status is
// automatically set to draft (or approved for standard changes, which
// are pre-approved by policy).
func (u *Usecase) Create(ctx context.Context, in CreateInput) (*model.ChangeRequest, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	now := u.clock.Now()
	status := model.StatusDraft
	// Standard changes are pre-approved -- skip the approval step.
	if in.Type == model.ChangeTypeStandard {
		status = model.StatusApproved
	}

	cr := &model.ChangeRequest{
		ID:             uuid.New().String(),
		Title:          strings.TrimSpace(in.Title),
		Description:    strings.TrimSpace(in.Description),
		Status:         status,
		Priority:       in.Priority,
		RiskLevel:      in.RiskLevel,
		Type:           in.Type,
		AssigneeID:     strings.TrimSpace(in.AssigneeID),
		CreatedBy:      strings.TrimSpace(in.CreatedBy),
		Implementation: strings.TrimSpace(in.Implementation),
		RollbackPlan:   strings.TrimSpace(in.RollbackPlan),
		ScheduledAt:    in.ScheduledAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := u.repo.Create(ctx, cr); err != nil {
		return nil, err
	}

	u.recordAudit(ctx, cr.ID, &model.AuditEntry{
		Timestamp: now,
		Action:    "created",
		Actor:     cr.CreatedBy,
		NewStatus: status,
	})

	u.log.Info("change request created",
		"id", cr.ID, "type", cr.Type, "priority", cr.Priority,
		"status", cr.Status, "created_by", cr.CreatedBy)

	return cr, nil
}

// Get returns a single ChangeRequest by ID.
func (u *Usecase) Get(ctx context.Context, id string) (*model.ChangeRequest, error) {
	return u.repo.Get(ctx, id)
}

// List returns change requests matching the filter, plus the total count
// before pagination.
func (u *Usecase) List(ctx context.Context, f ChangeFilter) ([]*model.ChangeRequest, int64, error) {
	total, err := u.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := u.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Update applies partial field changes to a draft or approved change.
// Status-transition methods (Approve, Reject, etc.) must be used for
// lifecycle changes.
func (u *Usecase) Update(ctx context.Context, id string, in UpdateInput) (*model.ChangeRequest, error) {
	cr, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if cr.Status != model.StatusDraft && cr.Status != model.StatusApproved {
		return nil, fmt.Errorf(
			"%w: cannot update change in %q status; only draft or approved changes are editable",
			errs.ErrInvalid, cr.Status)
	}

	if in.Title != nil {
		cr.Title = strings.TrimSpace(*in.Title)
	}
	if in.Description != nil {
		cr.Description = strings.TrimSpace(*in.Description)
	}
	if in.Priority != nil {
		v := strings.TrimSpace(*in.Priority)
		if !model.IsValidPriority(v) {
			return nil, fmt.Errorf("%w: invalid priority %q", errs.ErrInvalid, v)
		}
		cr.Priority = v
	}
	if in.RiskLevel != nil {
		v := strings.TrimSpace(*in.RiskLevel)
		if !model.IsValidRiskLevel(v) {
			return nil, fmt.Errorf("%w: invalid risk_level %q", errs.ErrInvalid, v)
		}
		cr.RiskLevel = v
	}
	if in.AssigneeID != nil {
		cr.AssigneeID = strings.TrimSpace(*in.AssigneeID)
	}
	if in.Implementation != nil {
		cr.Implementation = strings.TrimSpace(*in.Implementation)
	}
	if in.RollbackPlan != nil {
		cr.RollbackPlan = strings.TrimSpace(*in.RollbackPlan)
	}
	if in.ScheduledAt != nil {
		cr.ScheduledAt = in.ScheduledAt
	}
	cr.UpdatedAt = u.clock.Now()

	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, err
	}

	u.recordAudit(ctx, cr.ID, &model.AuditEntry{
		Timestamp: cr.UpdatedAt,
		Action:    "updated",
		Actor:     cr.CreatedBy,
	})

	return cr, nil
}

// Approve transitions a pending_approval change to approved.
func (u *Usecase) Approve(ctx context.Context, id, approverID string) (*model.ChangeRequest, error) {
	if strings.TrimSpace(approverID) == "" {
		return nil, fmt.Errorf("%w: approver_id is required", errs.ErrInvalid)
	}
	cr, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cr.ApproverID = strings.TrimSpace(approverID)
	cr.UpdatedAt = u.clock.Now()

	if err := u.doTransition(cr, model.StatusApproved); err != nil {
		return nil, err
	}
	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, err
	}
	u.recordAudit(ctx, cr.ID, &model.AuditEntry{
		Timestamp: cr.UpdatedAt,
		Action:    "approved",
		Actor:     approverID,
		OldStatus: model.StatusPendingApproval,
		NewStatus: model.StatusApproved,
	})
	return cr, nil
}

// Reject transitions a pending_approval change to rejected.
func (u *Usecase) Reject(ctx context.Context, id, approverID, reason string) (*model.ChangeRequest, error) {
	if strings.TrimSpace(approverID) == "" {
		return nil, fmt.Errorf("%w: approver_id is required", errs.ErrInvalid)
	}
	cr, err := u.transition(ctx, id, model.StatusRejected, approverID, "rejected")
	if err != nil {
		return nil, err
	}
	if reason != "" {
		u.recordAudit(ctx, id, &model.AuditEntry{
			Timestamp:   u.clock.Now(),
			Action:      "reject_reason",
			Actor:       approverID,
			Description: strings.TrimSpace(reason),
		})
	}
	return cr, nil
}

// Implement transitions an approved change to implementing.
func (u *Usecase) Implement(ctx context.Context, id, assigneeID string) (*model.ChangeRequest, error) {
	if strings.TrimSpace(assigneeID) == "" {
		return nil, fmt.Errorf("%w: assignee_id is required", errs.ErrInvalid)
	}
	cr, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cr.AssigneeID = strings.TrimSpace(assigneeID)
	cr.UpdatedAt = u.clock.Now()

	if err := u.doTransition(cr, model.StatusImplementing); err != nil {
		return nil, err
	}
	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, err
	}
	u.recordAudit(ctx, cr.ID, &model.AuditEntry{
		Timestamp: u.clock.Now(),
		Action:    "implementing",
		Actor:     assigneeID,
		OldStatus: model.StatusApproved,
		NewStatus: model.StatusImplementing,
	})
	return cr, nil
}

// Complete transitions an implementing change to completed.
func (u *Usecase) Complete(ctx context.Context, id, actor string) (*model.ChangeRequest, error) {
	if strings.TrimSpace(actor) == "" {
		return nil, fmt.Errorf("%w: actor is required", errs.ErrInvalid)
	}
	return u.transition(ctx, id, model.StatusCompleted, actor, "completed")
}

// Cancel transitions a change to cancelled from any cancellable state.
func (u *Usecase) Cancel(ctx context.Context, id, actor, reason string) (*model.ChangeRequest, error) {
	if strings.TrimSpace(actor) == "" {
		return nil, fmt.Errorf("%w: actor is required", errs.ErrInvalid)
	}
	cr, err := u.transition(ctx, id, model.StatusCancelled, actor, "cancelled")
	if err != nil {
		return nil, err
	}
	if reason != "" {
		u.recordAudit(ctx, id, &model.AuditEntry{
			Timestamp:   u.clock.Now(),
			Action:      "cancel_reason",
			Actor:       actor,
			Description: strings.TrimSpace(reason),
		})
	}
	return cr, nil
}

// GetAuditLog returns the full audit trail for a change request.
func (u *Usecase) GetAuditLog(ctx context.Context, id string) ([]*model.AuditEntry, error) {
	// Verify the change exists first.
	if _, err := u.repo.Get(ctx, id); err != nil {
		return nil, err
	}
	return u.audit.List(ctx, id)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// transition is the generic status-change path for simple transitions
// that only need an actor.
func (u *Usecase) transition(ctx context.Context, id, newStatus, actor, action string) (*model.ChangeRequest, error) {
	cr, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	oldStatus := cr.Status

	if err := u.doTransition(cr, newStatus); err != nil {
		return nil, err
	}
	cr.UpdatedAt = u.clock.Now()

	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, err
	}

	u.recordAudit(ctx, cr.ID, &model.AuditEntry{
		Timestamp: cr.UpdatedAt,
		Action:    action,
		Actor:     actor,
		OldStatus: oldStatus,
		NewStatus: newStatus,
	})

	u.log.Info("change status transition",
		"id", cr.ID, "from", oldStatus, "to", newStatus, "actor", actor)

	return cr, nil
}

// doTransition applies the status change on the struct and validates the
// transition is allowed. Does NOT persist -- caller is responsible.
func (u *Usecase) doTransition(cr *model.ChangeRequest, newStatus string) error {
	if !model.CanTransitionTo(cr.Status, newStatus) {
		return fmt.Errorf(
			"%w: cannot transition from %q to %q",
			errs.ErrInvalid, cr.Status, newStatus)
	}
	cr.Status = newStatus
	return nil
}

// recordAudit writes an audit entry. Failures are warn-logged but never
// returned -- audit must not block business (mirrors the project-wide
// audit philosophy in biz/audit).
func (u *Usecase) recordAudit(ctx context.Context, changeID string, entry *model.AuditEntry) {
	if u.audit == nil {
		return
	}
	if err := u.audit.Append(ctx, changeID, entry); err != nil {
		u.log.Warn("change audit write failed",
			"change_id", changeID, "action", entry.Action, "error", err)
	}
}

// validateCreateInput checks required fields and enum validity.
func validateCreateInput(in CreateInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("%w: title is required", errs.ErrInvalid)
	}
	if strings.TrimSpace(in.CreatedBy) == "" {
		return fmt.Errorf("%w: created_by is required", errs.ErrInvalid)
	}
	if in.Priority != "" && !model.IsValidPriority(in.Priority) {
		return fmt.Errorf("%w: invalid priority %q", errs.ErrInvalid, in.Priority)
	}
	if in.RiskLevel != "" && !model.IsValidRiskLevel(in.RiskLevel) {
		return fmt.Errorf("%w: invalid risk_level %q", errs.ErrInvalid, in.RiskLevel)
	}
	if in.Type != "" && !model.IsValidChangeType(in.Type) {
		return fmt.Errorf("%w: invalid type %q", errs.ErrInvalid, in.Type)
	}
	return nil
}
