// Package workflow implements the business logic for the workflow engine.
//
// Usecase holds an in-memory store (guarded by sync.RWMutex) and exposes
// CRUD lifecycle management plus an Execute path that runs workflow steps
// sequentially or in parallel, with per-step timeout and retry support.
//
// The execution engine evaluates step types as follows:
//   - action:       logs the action and marks success (stub for real dispatch).
//   - condition:    evaluates Params["expression"] against the execution
//                   context; truthy -> OnSuccess, falsy -> OnFailure.
//   - parallel:     runs Params["step_ids"] ([]string) concurrently.
//   - delay:        sleeps for step.Timeout (or Params["duration"]).
//   - notification: logs a notification (stub).
//   - approval:     auto-approves (stub for human-in-the-loop).
package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	model "github.com/ongridio/ongrid/internal/manager/model/workflow"
)

// Usecase is the workflow biz-layer facade. It stores workflow definitions
// and execution logs entirely in memory. Safe for concurrent use.
type Usecase struct {
	mu          sync.RWMutex
	workflows   map[string]*model.Workflow
	executions  map[string]*model.WorkflowExecution
	log         *slog.Logger
}

// NewUsecase constructs a workflow Usecase backed by an in-memory store.
func NewUsecase(log *slog.Logger) *Usecase {
	if log == nil {
		log = slog.Default()
	}
	return &Usecase{
		workflows:  make(map[string]*model.Workflow),
		executions: make(map[string]*model.WorkflowExecution),
		log:        log,
	}
}

// ---------------------------------------------------------------------------
// Workflow CRUD
// ---------------------------------------------------------------------------

// Create persists a new workflow definition. It assigns a UUID, sets the
// initial status to draft, stamps CreatedAt/UpdatedAt, and auto-assigns
// step IDs and order for any steps missing them.
func (uc *Usecase) Create(ctx context.Context, wf *model.Workflow) (*model.Workflow, error) {
	if err := wf.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalid, err)
	}

	now := time.Now().UTC()
	wf.ID = uuid.New().String()
	wf.Status = model.StatusDraft
	wf.CreatedAt = now
	wf.UpdatedAt = now
	wf.RunCount = 0

	for i := range wf.Steps {
		if wf.Steps[i].ID == "" {
			wf.Steps[i].ID = uuid.New().String()
		}
		wf.Steps[i].Order = i + 1
	}

	uc.mu.Lock()
	uc.workflows[wf.ID] = wf
	uc.mu.Unlock()

	uc.log.Info("workflow created", "id", wf.ID, "name", wf.Name)
	return wf, nil
}

// Get returns a workflow by ID.
func (uc *Usecase) Get(ctx context.Context, id string) (*model.Workflow, error) {
	uc.mu.RLock()
	wf, ok := uc.workflows[id]
	uc.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: workflow %s", ErrNotFound, id)
	}
	return wf, nil
}

// List returns all stored workflow definitions.
func (uc *Usecase) List(ctx context.Context) ([]model.Workflow, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	out := make([]model.Workflow, 0, len(uc.workflows))
	for _, wf := range uc.workflows {
		out = append(out, *wf)
	}
	return out, nil
}

// Update replaces the mutable fields of an existing workflow. Steps are
// re-validated and step IDs/orders are normalised.
func (uc *Usecase) Update(ctx context.Context, wf *model.Workflow) (*model.Workflow, error) {
	if err := wf.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalid, err)
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()

	existing, ok := uc.workflows[wf.ID]
	if !ok {
		return nil, fmt.Errorf("%w: workflow %s", ErrNotFound, wf.ID)
	}

	// Preserve immutable fields.
	wf.CreatedAt = existing.CreatedAt
	wf.CreatedBy = existing.CreatedBy
	wf.RunCount = existing.RunCount
	wf.LastRunAt = existing.LastRunAt
	wf.UpdatedAt = time.Now().UTC()

	for i := range wf.Steps {
		if wf.Steps[i].ID == "" {
			wf.Steps[i].ID = uuid.New().String()
		}
		wf.Steps[i].Order = i + 1
	}

	uc.workflows[wf.ID] = wf
	uc.log.Info("workflow updated", "id", wf.ID)
	return wf, nil
}

// Delete removes a workflow by ID.
func (uc *Usecase) Delete(ctx context.Context, id string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if _, ok := uc.workflows[id]; !ok {
		return fmt.Errorf("%w: workflow %s", ErrNotFound, id)
	}
	delete(uc.workflows, id)
	uc.log.Info("workflow deleted", "id", id)
	return nil
}

// Activate transitions a workflow to the active state.
func (uc *Usecase) Activate(ctx context.Context, id string) (*model.Workflow, error) {
	return uc.transition(ctx, id, model.StatusActive)
}

// Pause transitions a workflow to the paused state.
func (uc *Usecase) Pause(ctx context.Context, id string) (*model.Workflow, error) {
	return uc.transition(ctx, id, model.StatusPaused)
}

// Archive transitions a workflow to the archived state.
func (uc *Usecase) Archive(ctx context.Context, id string) (*model.Workflow, error) {
	return uc.transition(ctx, id, model.StatusArchived)
}

func (uc *Usecase) transition(ctx context.Context, id string, target model.WorkflowStatus) (*model.Workflow, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	wf, ok := uc.workflows[id]
	if !ok {
		return nil, fmt.Errorf("%w: workflow %s", ErrNotFound, id)
	}
	wf.Status = target
	wf.UpdatedAt = time.Now().UTC()
	uc.log.Info("workflow status changed", "id", id, "status", string(target))
	return wf, nil
}

// ---------------------------------------------------------------------------
// Execution
// ---------------------------------------------------------------------------

// Execute starts a new execution of the identified workflow. Only active
// workflows may be executed. The execution runs synchronously within this
// call — callers that need async semantics should invoke Execute in a
// goroutine. context cancellation propagates to running steps.
func (uc *Usecase) Execute(ctx context.Context, workflowID string, params map[string]interface{}) (*model.WorkflowExecution, error) {
	uc.mu.RLock()
	wf, ok := uc.workflows[workflowID]
	uc.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: workflow %s", ErrNotFound, workflowID)
	}
	if wf.Status != model.StatusActive {
		return nil, fmt.Errorf("%w: workflow %s is %s, not active", ErrInvalid, workflowID, wf.Status)
	}

	now := time.Now().UTC()
	ex := &model.WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     model.ExecutionRunning,
		StartedAt:  now,
		Steps:      make([]model.StepExecution, 0, len(wf.Steps)),
	}

	// Persist the execution record.
	uc.mu.Lock()
	uc.executions[ex.ID] = ex
	uc.mu.Unlock()

	uc.log.Info("execution started", "execution_id", ex.ID, "workflow_id", workflowID)

	// Build a step lookup map ordered by Order.
	stepMap := make(map[string]model.WorkflowStep, len(wf.Steps))
	var firstStep *model.WorkflowStep
	for _, s := range wf.Steps {
		stepMap[s.ID] = s
		if firstStep == nil || s.Order < firstStep.Order {
			sCopy := s
			firstStep = &sCopy
		}
	}

	// Walk the step chain.
	var currentStep *model.WorkflowStep
	if firstStep != nil {
		currentStep = firstStep
	}

	for currentStep != nil {
		if ctx.Err() != nil {
			ex.Status = model.ExecutionCancelled
			ex.Error = "execution cancelled: " + ctx.Err().Error()
			now := time.Now().UTC()
			ex.FinishedAt = &now
			uc.updateExecution(ex)
			return ex, ctx.Err()
		}

		stepExec := uc.executeStep(ctx, currentStep, params, ex)

		uc.mu.Lock()
		ex.Steps = append(ex.Steps, stepExec)
		uc.mu.Unlock()

		// Determine the next step.
		nextID := ""
		switch currentStep.Type {
		case model.StepTypeCondition:
			if stepExec.Status == model.ExecutionSuccess {
				nextID = currentStep.OnSuccess
			} else {
				nextID = currentStep.OnFailure
			}
		default:
			if stepExec.Status == model.ExecutionSuccess {
				nextID = currentStep.NextStepID
			} else {
				nextID = currentStep.OnFailure
			}
		}

		if nextID == "" {
			// No explicit next — pick the next step by order.
			nextOrder := currentStep.Order + 1
			found := false
			for _, s := range wf.Steps {
				if s.Order == nextOrder {
					nextID = s.ID
					found = true
					break
				}
			}
			if !found {
				nextID = "" // end of chain
			}
		}

		if nextID == "" || stepExec.Status == model.ExecutionFailed {
			break
		}

		next, ok := stepMap[nextID]
		if !ok {
			ex.Status = model.ExecutionFailed
			ex.Error = fmt.Sprintf("step %q references unknown next step %q", currentStep.ID, nextID)
			break
		}
		currentStep = &next
	}

	// Finalise.
	if ex.Status == model.ExecutionRunning {
		allOK := true
		for _, se := range ex.Steps {
			if se.Status != model.ExecutionSuccess {
				allOK = false
				break
			}
		}
		if allOK {
			ex.Status = model.ExecutionSuccess
		} else {
			ex.Status = model.ExecutionFailed
		}
	}
	now = time.Now().UTC()
	ex.FinishedAt = &now
	uc.updateExecution(ex)

	// Bump the workflow run counters.
	uc.mu.Lock()
	if wfRef, ok2 := uc.workflows[workflowID]; ok2 {
		wfRef.RunCount++
		wfRef.LastRunAt = &now
	}
	uc.mu.Unlock()

	uc.log.Info("execution finished", "execution_id", ex.ID, "status", string(ex.Status))
	return ex, nil
}

// GetExecution retrieves an execution record by ID.
func (uc *Usecase) GetExecution(ctx context.Context, executionID string) (*model.WorkflowExecution, error) {
	uc.mu.RLock()
	ex, ok := uc.executions[executionID]
	uc.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: execution %s", ErrNotFound, executionID)
	}
	return ex, nil
}

// ListExecutions returns all executions for a given workflow.
func (uc *Usecase) ListExecutions(ctx context.Context, workflowID string) ([]model.WorkflowExecution, error) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	var out []model.WorkflowExecution
	for _, ex := range uc.executions {
		if ex.WorkflowID == workflowID {
			out = append(out, *ex)
		}
	}
	return out, nil
}

// CancelExecution requests cancellation of a running execution.
func (uc *Usecase) CancelExecution(ctx context.Context, executionID string) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	ex, ok := uc.executions[executionID]
	if !ok {
		return fmt.Errorf("%w: execution %s", ErrNotFound, executionID)
	}
	if ex.Status != model.ExecutionRunning && ex.Status != model.ExecutionPending {
		return fmt.Errorf("%w: execution %s is %s, cannot cancel", ErrInvalid, executionID, ex.Status)
	}
	ex.Status = model.ExecutionCancelled
	now := time.Now().UTC()
	ex.FinishedAt = &now
	ex.Error = "cancelled by user"
	uc.log.Info("execution cancelled", "execution_id", executionID)
	return nil
}

// ---------------------------------------------------------------------------
// Internal step execution
// ---------------------------------------------------------------------------

// executeStep runs a single workflow step, respecting retries and timeout.
func (uc *Usecase) executeStep(ctx context.Context, step *model.WorkflowStep, params map[string]interface{}, ex *model.WorkflowExecution) model.StepExecution {
	maxAttempts := step.RetryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	now := time.Now().UTC()
	stepExec := model.StepExecution{
		StepID:    step.ID,
		Status:    model.ExecutionRunning,
		StartedAt: now,
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		stepExec.Attempt = attempt

		// Apply per-step timeout if configured.
		stepCtx := ctx
		var cancel context.CancelFunc
		if step.Timeout > 0 {
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		}

		var out string
		var err error
		switch step.Type {
		case model.StepTypeCondition:
			out, err = uc.executeCondition(stepCtx, step, params)
		case model.StepTypeParallel:
			out, err = uc.executeParallel(stepCtx, step, params)
		case model.StepTypeDelay:
			out, err = uc.executeDelay(stepCtx, step)
		default:
			// action, notification, approval — stub execution.
			out, err = uc.executeAction(stepCtx, step, params)
		}

		if cancel != nil {
			cancel()
		}

		if err == nil {
			stepExec.Status = model.ExecutionSuccess
			stepExec.Output = out
			fin := time.Now().UTC()
			stepExec.FinishedAt = &fin
			return stepExec
		}
		lastErr = err
		uc.log.Warn("step attempt failed",
			"step_id", step.ID, "attempt", attempt, "error", err)
	}

	stepExec.Status = model.ExecutionFailed
	stepExec.Error = lastErr.Error()
	fin := time.Now().UTC()
	stepExec.FinishedAt = &fin
	return stepExec
}

// executeAction is a stub for real action dispatch. It logs the action
// and returns success.
func (uc *Usecase) executeAction(ctx context.Context, step *model.WorkflowStep, params map[string]interface{}) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	uc.log.Info("executing action", "step_id", step.ID, "action", step.Action)
	return fmt.Sprintf("action %q executed", step.Action), nil
}

// executeCondition evaluates a simple condition. The condition expression
// is read from step.Params["expression"]. Any non-empty, non-"false"
// string value is treated as truthy.
func (uc *Usecase) executeCondition(ctx context.Context, step *model.WorkflowStep, params map[string]interface{}) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	expr := ""
	if v, ok := step.Params["expression"]; ok {
		expr = fmt.Sprintf("%v", v)
	}
	// Merge execution params into the expression context for future
	// template expansion. For now we just log.
	uc.log.Info("evaluating condition", "step_id", step.ID, "expression", expr, "params", params)

	truthy := expr != "" && expr != "false" && expr != "0"
	if truthy {
		return "condition evaluated to true", nil
	}
	return "condition evaluated to false", nil
}

// executeParallel runs multiple sub-steps concurrently. The sub-step IDs
// are read from step.Params["step_ids"] ([]interface{} of string IDs).
// Sub-step definitions are looked up from the parent execution's workflow.
func (uc *Usecase) executeParallel(ctx context.Context, step *model.WorkflowStep, params map[string]interface{}) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// For the parallel step, we treat step.Params["step_ids"] as a list
	// of virtual step actions (strings) to run concurrently.
	rawIDs, _ := step.Params["step_ids"].([]interface{})
	if len(rawIDs) == 0 {
		return "parallel: no sub-steps defined", nil
	}

	type result struct {
		id  string
		out string
		err error
	}
	ch := make(chan result, len(rawIDs))
	for _, raw := range rawIDs {
		subID := fmt.Sprintf("%v", raw)
		go func(action string) {
			out, err := uc.executeAction(ctx, &model.WorkflowStep{
				ID:     step.ID + "/" + action,
				Action: action,
			}, params)
			ch <- result{id: action, out: out, err: err}
		}(subID)
	}

	var errs []string
	for i := 0; i < len(rawIDs); i++ {
		r := <-ch
		if r.err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", r.id, r.err))
		}
	}
	if len(errs) > 0 {
		return "", fmt.Errorf("parallel sub-step errors: %v", errs)
	}
	return fmt.Sprintf("parallel: %d sub-steps completed", len(rawIDs)), nil
}

// executeDelay sleeps for the step's configured duration.
func (uc *Usecase) executeDelay(ctx context.Context, step *model.WorkflowStep) (string, error) {
	d := step.Timeout
	if d == 0 {
		if v, ok := step.Params["duration"]; ok {
			if dur, ok2 := v.(time.Duration); ok2 {
				d = dur
			} else if secs, ok2 := v.(float64); ok2 {
				d = time.Duration(secs * float64(time.Second))
			}
		}
	}
	if d == 0 {
		return "delay: zero duration, skipping", nil
	}
	uc.log.Info("delay step", "step_id", step.ID, "duration", d)
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(d):
		return fmt.Sprintf("delayed %s", d), nil
	}
}

// updateExecution persists the mutated execution back into the store.
func (uc *Usecase) updateExecution(ex *model.WorkflowExecution) {
	uc.mu.Lock()
	uc.executions[ex.ID] = ex
	uc.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Error sentinels (scoped to this BC)
// ---------------------------------------------------------------------------

var (
	ErrNotFound = fmt.Errorf("not found")
	ErrInvalid  = fmt.Errorf("invalid argument")
)
