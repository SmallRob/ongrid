// Package workflow holds domain entities for the manager/workflow sub-domain.
// A workflow is a user-defined sequence of steps that the engine executes
// either on-demand (manual), by cron schedule, by webhook call, or when an
// alert fires. Each execution produces a WorkflowExecution record that
// captures per-step outcomes for observability and retry.
package workflow

import (
	"fmt"
	"time"
)

// StepType enumerates the kinds of steps a workflow can contain.
type StepType string

const (
	StepTypeAction      StepType = "action"
	StepTypeCondition   StepType = "condition"
	StepTypeParallel    StepType = "parallel"
	StepTypeDelay       StepType = "delay"
	StepTypeNotification StepType = "notification"
	StepTypeApproval    StepType = "approval"
)

// IsKnownStepType reports whether t is a recognised step type.
func IsKnownStepType(t StepType) bool {
	switch t {
	case StepTypeAction, StepTypeCondition, StepTypeParallel,
		StepTypeDelay, StepTypeNotification, StepTypeApproval:
		return true
	}
	return false
}

// TriggerType enumerates how a workflow can be invoked.
type TriggerType string

const (
	TriggerManual   TriggerType = "manual"
	TriggerCron     TriggerType = "cron"
	TriggerWebhook  TriggerType = "webhook"
	TriggerAlert    TriggerType = "alert"
)

// WorkflowStatus represents the lifecycle state of a workflow definition.
type WorkflowStatus string

const (
	StatusDraft    WorkflowStatus = "draft"
	StatusActive   WorkflowStatus = "active"
	StatusPaused   WorkflowStatus = "paused"
	StatusArchived WorkflowStatus = "archived"
)

// ExecutionStatus represents the lifecycle state of a single execution run.
type ExecutionStatus string

const (
	ExecutionPending   ExecutionStatus = "pending"
	ExecutionRunning   ExecutionStatus = "running"
	ExecutionSuccess   ExecutionStatus = "success"
	ExecutionFailed    ExecutionStatus = "failed"
	ExecutionCancelled ExecutionStatus = "cancelled"
	ExecutionTimeout   ExecutionStatus = "timeout"
)

// Workflow is the top-level definition that users create and manage.
type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
	Status      WorkflowStatus `json:"status"`
	Trigger     WorkflowTrigger `json:"trigger"`
	CreatedBy   string         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastRunAt   *time.Time     `json:"last_run_at,omitempty"`
	RunCount    int            `json:"run_count"`
}

// WorkflowStep describes a single node in the workflow DAG.
type WorkflowStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       StepType               `json:"type"`
	Action     string                 `json:"action,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
	NextStepID string                 `json:"next_step_id,omitempty"`
	OnSuccess  string                 `json:"on_success,omitempty"`
	OnFailure  string                 `json:"on_failure,omitempty"`
	Timeout    time.Duration          `json:"timeout"`
	RetryCount int                    `json:"retry_count"`
	Order      int                    `json:"order"`
}

// WorkflowTrigger configures how a workflow is activated.
type WorkflowTrigger struct {
	Type         TriggerType `json:"type"`
	CronExpr     string      `json:"cron_expr,omitempty"`
	WebhookPath  string      `json:"webhook_path,omitempty"`
	Manual       bool        `json:"manual"`
}

// WorkflowExecution records a single run of a workflow.
type WorkflowExecution struct {
	ID         string          `json:"id"`
	WorkflowID string          `json:"workflow_id"`
	Status     ExecutionStatus `json:"status"`
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at,omitempty"`
	Steps      []StepExecution `json:"steps"`
	Error      string          `json:"error,omitempty"`
}

// StepExecution records the outcome of a single step within an execution.
type StepExecution struct {
	StepID    string          `json:"step_id"`
	Status    ExecutionStatus `json:"status"`
	StartedAt time.Time       `json:"started_at"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	Output    string          `json:"output,omitempty"`
	Error     string          `json:"error,omitempty"`
	Attempt   int             `json:"attempt"`
}

// Validate checks required fields on the workflow before persistence.
func (w *Workflow) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if len(w.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}
	for i, s := range w.Steps {
		if s.Name == "" {
			return fmt.Errorf("step[%d]: name is required", i)
		}
		if !IsKnownStepType(s.Type) {
			return fmt.Errorf("step[%d]: unknown type %q", i, s.Type)
		}
	}
	return nil
}
