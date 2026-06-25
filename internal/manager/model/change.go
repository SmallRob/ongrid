// Package change holds domain entities for the ITSM change management
// sub-domain. A ChangeRequest models the full lifecycle of an IT change
// from draft through approval, implementation and completion (or
// cancellation). The model is designed after the ITIL v4 change enablement
// practice with three change types (standard, normal, emergency) and
// explicit risk assessment.
package change

import "time"

// Change status constants. Transitions are guarded by the biz layer;
// the model only declares the allowed values.
const (
	StatusDraft          = "draft"
	StatusPendingApproval = "pending_approval"
	StatusApproved       = "approved"
	StatusRejected       = "rejected"
	StatusImplementing   = "implementing"
	StatusCompleted      = "completed"
	StatusCancelled      = "cancelled"
)

// Priority levels for a change request. Drives scheduling urgency and
// notification escalation.
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Risk levels assigned during change assessment. Used by the approval
// workflow to determine the required approver tier.
const (
	RiskLevelLow    = "low"
	RiskLevelMedium = "medium"
	RiskLevelHigh   = "high"
)

// ChangeType follows ITIL classification. Standard changes are
// pre-approved; normal changes require CAB review; emergency changes
// bypass the normal approval window but require post-implementation
// review.
const (
	ChangeTypeStandard  = "standard"
	ChangeTypeNormal    = "normal"
	ChangeTypeEmergency = "emergency"
)

// ChangeRequest is the aggregate root for the change management bounded
// context. All mutable fields are written atomically through the usecase;
// callers must not modify the struct directly.
type ChangeRequest struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	Priority        string     `json:"priority"`
	RiskLevel       string     `json:"risk_level"`
	Type            string     `json:"type"`
	ApproverID      string     `json:"approver_id,omitempty"`
	AssigneeID      string     `json:"assignee_id,omitempty"`
	CreatedBy       string     `json:"created_by"`
	Implementation  string     `json:"implementation,omitempty"`
	RollbackPlan    string     `json:"rollback_plan,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ValidStatuses returns every allowed status string. Useful for input
// validation without importing the biz layer.
var ValidStatuses = []string{
	StatusDraft, StatusPendingApproval, StatusApproved,
	StatusRejected, StatusImplementing, StatusCompleted, StatusCancelled,
}

// ValidPriorities returns every allowed priority string.
var ValidPriorities = []string{
	PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical,
}

// ValidRiskLevels returns every allowed risk level string.
var ValidRiskLevels = []string{
	RiskLevelLow, RiskLevelMedium, RiskLevelHigh,
}

// ValidChangeTypes returns every allowed change type string.
var ValidChangeTypes = []string{
	ChangeTypeStandard, ChangeTypeNormal, ChangeTypeEmergency,
}

// ValidStatusTransitions defines the allowed next-status set from each
// current status. The biz layer enforces these; the model only holds
// the table so both handler and usecase can reference it.
var ValidStatusTransitions = map[string][]string{
	StatusDraft:          {StatusPendingApproval, StatusCancelled},
	StatusPendingApproval: {StatusApproved, StatusRejected, StatusCancelled},
	StatusApproved:       {StatusImplementing, StatusCancelled},
	StatusRejected:       {StatusDraft},
	StatusImplementing:   {StatusCompleted, StatusCancelled},
	StatusCompleted:      {},
	StatusCancelled:      {StatusDraft},
}

// IsValidStatus reports whether s is a recognized change status.
func IsValidStatus(s string) bool {
	for _, v := range ValidStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// IsValidPriority reports whether p is a recognized priority level.
func IsValidPriority(p string) bool {
	for _, v := range ValidPriorities {
		if v == p {
			return true
		}
	}
	return false
}

// IsValidRiskLevel reports whether r is a recognized risk level.
func IsValidRiskLevel(r string) bool {
	for _, v := range ValidRiskLevels {
		if v == r {
			return true
		}
	}
	return false
}

// IsValidChangeType reports whether t is a recognized change type.
func IsValidChangeType(t string) bool {
	for _, v := range ValidChangeTypes {
		if v == t {
			return true
		}
	}
	return false
}

// CanTransitionTo reports whether a transition from current status to
// next is allowed by the lifecycle rules.
func CanTransitionTo(current, next string) bool {
	allowed, ok := ValidStatusTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// AuditEntry records a single state transition or significant event on
// a ChangeRequest. The biz layer appends entries so the handler layer
// can surface an audit trail without querying a separate log table.
type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	Actor       string    `json:"actor"`
	OldStatus   string    `json:"old_status,omitempty"`
	NewStatus   string    `json:"new_status,omitempty"`
	Description string    `json:"description,omitempty"`
}
