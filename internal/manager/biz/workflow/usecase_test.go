package workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	model "github.com/ongridio/ongrid/internal/manager/model/workflow"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestUsecase(t *testing.T) *Usecase {
	t.Helper()
	return NewUsecase(nil)
}

func sampleWorkflow() *model.Workflow {
	return &model.Workflow{
		Name:        "test-workflow",
		Description: "A simple test workflow",
		Steps: []model.WorkflowStep{
			{
				Name:   "step-1",
				Type:   model.StepTypeAction,
				Action: "echo",
			},
			{
				Name:   "step-2",
				Type:   model.StepTypeAction,
				Action: "notify",
			},
		},
		Trigger: model.WorkflowTrigger{
			Type:   model.TriggerManual,
			Manual: true,
		},
		CreatedBy: "tester",
	}
}

// ---------------------------------------------------------------------------
// CRUD tests
// ---------------------------------------------------------------------------

func TestCreate_Get_List(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, err := uc.Create(ctx, sampleWorkflow())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if wf.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if wf.Status != model.StatusDraft {
		t.Fatalf("expected status draft, got %s", wf.Status)
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(wf.Steps))
	}
	// Steps should have auto-assigned IDs and orders.
	for i, s := range wf.Steps {
		if s.ID == "" {
			t.Fatalf("step[%d]: expected non-empty ID", i)
		}
		if s.Order != i+1 {
			t.Fatalf("step[%d]: expected order %d, got %d", i, i+1, s.Order)
		}
	}

	// Get.
	got, err := uc.Get(ctx, wf.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "test-workflow" {
		t.Fatalf("expected name test-workflow, got %s", got.Name)
	}

	// List.
	all, err := uc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(all))
	}
}

func TestCreate_Invalid(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	// Missing name.
	_, err := uc.Create(ctx, &model.Workflow{
		Steps: []model.WorkflowStep{{Name: "s", Type: model.StepTypeAction}},
	})
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	// Missing steps.
	_, err = uc.Create(ctx, &model.Workflow{Name: "no-steps"})
	if err == nil {
		t.Fatal("expected error for missing steps")
	}

	// Unknown step type.
	_, err = uc.Create(ctx, &model.Workflow{
		Name:  "bad-type",
		Steps: []model.WorkflowStep{{Name: "s", Type: "bogus"}},
	})
	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
}

func TestUpdate(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	wf.Name = "updated-name"
	updated, err := uc.Update(ctx, wf)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "updated-name" {
		t.Fatalf("expected updated-name, got %s", updated.Name)
	}
	// CreatedAt should be preserved.
	if !updated.CreatedAt.Equal(wf.CreatedAt) {
		t.Fatal("expected CreatedAt to be preserved")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	_, err := uc.Update(ctx, &model.Workflow{
		ID:   "nonexistent",
		Name: "x",
		Steps: []model.WorkflowStep{
			{Name: "s", Type: model.StepTypeAction},
		},
	})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestDelete(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	if err := uc.Delete(ctx, wf.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := uc.Get(ctx, wf.ID)
	if err == nil {
		t.Fatal("expected not-found after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	err := uc.Delete(ctx, "nope")
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

// ---------------------------------------------------------------------------
// Lifecycle transitions
// ---------------------------------------------------------------------------

func TestActivate_Pause_Archive(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())

	// Activate.
	wf, err := uc.Activate(ctx, wf.ID)
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if wf.Status != model.StatusActive {
		t.Fatalf("expected active, got %s", wf.Status)
	}

	// Pause.
	wf, err = uc.Pause(ctx, wf.ID)
	if err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if wf.Status != model.StatusPaused {
		t.Fatalf("expected paused, got %s", wf.Status)
	}

	// Archive.
	wf, err = uc.Archive(ctx, wf.ID)
	if err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if wf.Status != model.StatusArchived {
		t.Fatalf("expected archived, got %s", wf.Status)
	}
}

func TestActivate_NotFound(t *testing.T) {
	uc := newTestUsecase(t)
	_, err := uc.Activate(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected not-found")
	}
}

// ---------------------------------------------------------------------------
// Execute tests
// ---------------------------------------------------------------------------

func TestExecute_SimpleWorkflow(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	uc.Activate(ctx, wf.ID)

	ex, err := uc.Execute(ctx, wf.ID, map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ex.Status != model.ExecutionSuccess {
		t.Fatalf("expected success, got %s", ex.Status)
	}
	if len(ex.Steps) != 2 {
		t.Fatalf("expected 2 step executions, got %d", len(ex.Steps))
	}
	for i, se := range ex.Steps {
		if se.Status != model.ExecutionSuccess {
			t.Fatalf("step[%d] expected success, got %s", i, se.Status)
		}
	}
	if ex.FinishedAt == nil {
		t.Fatal("expected FinishedAt to be set")
	}

	// Verify RunCount bumped.
	got, _ := uc.Get(ctx, wf.ID)
	if got.RunCount != 1 {
		t.Fatalf("expected RunCount 1, got %d", got.RunCount)
	}
	if got.LastRunAt == nil {
		t.Fatal("expected LastRunAt to be set")
	}
}

func TestExecute_ConditionBranch(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	// Condition evaluates to true -> OnSuccess step, false -> OnFailure step.
	successStep := model.WorkflowStep{
		ID:     "success-path",
		Name:   "success-path",
		Type:   model.StepTypeAction,
		Action: "ok",
		Order:  99,
	}
	failureStep := model.WorkflowStep{
		ID:     "failure-path",
		Name:   "failure-path",
		Type:   model.StepTypeAction,
		Action: "fail",
		Order:  99,
	}

	// Truthy condition.
	wf := &model.Workflow{
		Name: "condition-wf",
		Steps: []model.WorkflowStep{
			{
				Name:       "check",
				Type:       model.StepTypeCondition,
				Params:     map[string]interface{}{"expression": "true"},
				OnSuccess:  "success-path",
				OnFailure:  "failure-path",
			},
			successStep,
			failureStep,
		},
		Trigger:   model.WorkflowTrigger{Type: model.TriggerManual, Manual: true},
		CreatedBy: "tester",
	}

	created, _ := uc.Create(ctx, wf)
	uc.Activate(ctx, created.ID)

	ex, err := uc.Execute(ctx, created.ID, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	// Should go through the condition (true) -> success-path.
	if ex.Status != model.ExecutionSuccess {
		t.Fatalf("expected success, got %s (err=%s)", ex.Status, ex.Error)
	}

	// Falsy condition.
	wf2 := &model.Workflow{
		Name: "condition-wf-false",
		Steps: []model.WorkflowStep{
			{
				Name:       "check",
				Type:       model.StepTypeCondition,
				Params:     map[string]interface{}{"expression": "false"},
				OnSuccess:  "success-path",
				OnFailure:  "failure-path",
			},
			successStep,
			failureStep,
		},
		Trigger:   model.WorkflowTrigger{Type: model.TriggerManual, Manual: true},
		CreatedBy: "tester",
	}
	created2, _ := uc.Create(ctx, wf2)
	uc.Activate(ctx, created2.ID)

	ex2, err := uc.Execute(ctx, created2.ID, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ex2.Status != model.ExecutionSuccess {
		t.Fatalf("expected success (failure-path still succeeds), got %s", ex2.Status)
	}
}

func TestExecute_OnlyActiveWorkflows(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	// Not activated — still draft.
	_, err := uc.Execute(ctx, wf.ID, nil)
	if err == nil {
		t.Fatal("expected error for non-active workflow")
	}
}

func TestExecute_NotFound(t *testing.T) {
	uc := newTestUsecase(t)
	_, err := uc.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected not-found")
	}
}

// ---------------------------------------------------------------------------
// Execution CRUD tests
// ---------------------------------------------------------------------------

func TestGetExecution(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	uc.Activate(ctx, wf.ID)

	ex, _ := uc.Execute(ctx, wf.ID, nil)

	got, err := uc.GetExecution(ctx, ex.ID)
	if err != nil {
		t.Fatalf("GetExecution: %v", err)
	}
	if got.ID != ex.ID {
		t.Fatalf("expected execution %s, got %s", ex.ID, got.ID)
	}
}

func TestGetExecution_NotFound(t *testing.T) {
	uc := newTestUsecase(t)
	_, err := uc.GetExecution(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected not-found")
	}
}

func TestListExecutions(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	uc.Activate(ctx, wf.ID)

	uc.Execute(ctx, wf.ID, nil)
	uc.Execute(ctx, wf.ID, nil)

	execs, err := uc.ListExecutions(ctx, wf.ID)
	if err != nil {
		t.Fatalf("ListExecutions: %v", err)
	}
	if len(execs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(execs))
	}
}

// ---------------------------------------------------------------------------
// CancelExecution tests
// ---------------------------------------------------------------------------

func TestCancelExecution_CompletedExecution(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf, _ := uc.Create(ctx, sampleWorkflow())
	uc.Activate(ctx, wf.ID)

	ex, _ := uc.Execute(ctx, wf.ID, nil)

	// Already completed — should fail.
	err := uc.CancelExecution(ctx, ex.ID)
	if err == nil {
		t.Fatal("expected error cancelling completed execution")
	}
}

func TestCancelExecution_NotFound(t *testing.T) {
	uc := newTestUsecase(t)
	err := uc.CancelExecution(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected not-found")
	}
}

func TestCancelExecution_ViaContext(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	// Create a workflow with a delay step to give us time to cancel.
	wf := &model.Workflow{
		Name: "cancel-wf",
		Steps: []model.WorkflowStep{
			{
				Name:    "delay-step",
				Type:    model.StepTypeDelay,
				Timeout: 10 * time.Second,
			},
		},
		Trigger:   model.WorkflowTrigger{Type: model.TriggerManual, Manual: true},
		CreatedBy: "tester",
	}
	created, _ := uc.Create(ctx, wf)
	uc.Activate(ctx, created.ID)

	// Run execute in a goroutine with a cancellable context.
	execCtx, cancel := context.WithCancel(ctx)
	type execResult struct {
		ex  *model.WorkflowExecution
		err error
	}
	ch := make(chan execResult, 1)
	go func() {
		ex, err := uc.Execute(execCtx, created.ID, nil)
		ch <- execResult{ex: ex, err: err}
	}()

	// Give the execution a moment to start, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	res := <-ch
	if res.err == nil {
		t.Fatal("expected context cancellation error")
	}
	if res.ex != nil && res.ex.Status != model.ExecutionCancelled {
		t.Fatalf("expected cancelled status, got %s", res.ex.Status)
	}
}

// ---------------------------------------------------------------------------
// Parallel execution test
// ---------------------------------------------------------------------------

func TestExecute_ParallelStep(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	wf := &model.Workflow{
		Name: "parallel-wf",
		Steps: []model.WorkflowStep{
			{
				Name: "parallel-step",
				Type: model.StepTypeParallel,
				Params: map[string]interface{}{
					"step_ids": []interface{}{"task-a", "task-b", "task-c"},
				},
			},
		},
		Trigger:   model.WorkflowTrigger{Type: model.TriggerManual, Manual: true},
		CreatedBy: "tester",
	}
	created, _ := uc.Create(ctx, wf)
	uc.Activate(ctx, created.ID)

	ex, err := uc.Execute(ctx, created.ID, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ex.Status != model.ExecutionSuccess {
		t.Fatalf("expected success, got %s (err=%s)", ex.Status, ex.Error)
	}
}

// ---------------------------------------------------------------------------
// Error sentinels
// ---------------------------------------------------------------------------

func TestErrorSentinels(t *testing.T) {
	uc := newTestUsecase(t)
	ctx := context.Background()

	_, err := uc.Get(ctx, "nonexistent")
	if !isNotFound(err) {
		t.Fatalf("expected not-found, got %v", err)
	}

	_, err = uc.Create(ctx, &model.Workflow{})
	if !isInvalid(err) {
		t.Fatalf("expected invalid, got %v", err)
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func isInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}
