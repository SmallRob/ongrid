// Package change is the manager/change biz tier. It owns the lifecycle
// of ChangeRequest records: creation, approval workflow, implementation
// tracking and completion.
package change

import (
	"context"

	model "github.com/ongridio/ongrid/internal/manager/model/change"
)

// ChangeFilter narrows ChangeRequest.List results. Empty / zero-value
// fields mean "no filter". Limit/Offset use repo defaults when zero.
type ChangeFilter struct {
	Status   string
	Priority string
	Type     string
	Assignee string
	Creator  string
	Limit    int
	Offset   int
}

// Repo is the change persistence contract. The in-memory implementation
// lives in the same package; a future DB-backed implementation would
// live under internal/manager/data/change.
type Repo interface {
	// Create persists a new ChangeRequest. The caller is responsible for
	// generating the ID and timestamps. Returns ErrConflict if the ID
	// already exists.
	Create(ctx context.Context, cr *model.ChangeRequest) error

	// Get returns the ChangeRequest by ID; ErrNotFound otherwise.
	Get(ctx context.Context, id string) (*model.ChangeRequest, error)

	// List returns change requests matching f. Sorted by CreatedAt DESC
	// (newest first).
	List(ctx context.Context, f ChangeFilter) ([]*model.ChangeRequest, error)

	// Update replaces the mutable fields of an existing ChangeRequest.
	// Returns ErrNotFound if the ID does not exist.
	Update(ctx context.Context, cr *model.ChangeRequest) error

	// Count returns the total number of change requests.
	Count(ctx context.Context) (int64, error)
}

// AuditRepo is the persistence contract for change audit trail entries.
// Separated from the main Repo so audit writes can be fire-and-forget
// without blocking the business transaction.
type AuditRepo interface {
	// Append adds an AuditEntry for the given change request ID.
	Append(ctx context.Context, changeID string, entry *model.AuditEntry) error

	// List returns all audit entries for a change request, ordered by
	// timestamp ascending.
	List(ctx context.Context, changeID string) ([]*model.AuditEntry, error)
}
