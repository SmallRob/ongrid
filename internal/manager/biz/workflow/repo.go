package workflow

import (
	"context"

	model "github.com/ongridio/ongrid/internal/manager/model/workflow"
)

// Repo is the biz-layer persistence contract for the workflow sub-domain.
//
// Implementations may be backed by SQLite, PostgreSQL, or an in-memory
// store (as in this reference implementation). The interface mirrors
// the read/write split used across other Ongrid bounded contexts.
type Repo interface {
	// Workflow CRUD.
	Create(ctx context.Context, wf *model.Workflow) error
	Get(ctx context.Context, id string) (*model.Workflow, error)
	List(ctx context.Context) ([]model.Workflow, error)
	Update(ctx context.Context, wf *model.Workflow) error
	Delete(ctx context.Context, id string) error

	// Execution log.
	CreateExecution(ctx context.Context, ex *model.WorkflowExecution) error
	GetExecution(ctx context.Context, id string) (*model.WorkflowExecution, error)
	UpdateExecution(ctx context.Context, ex *model.WorkflowExecution) error
	ListExecutions(ctx context.Context, workflowID string) ([]model.WorkflowExecution, error)
}
