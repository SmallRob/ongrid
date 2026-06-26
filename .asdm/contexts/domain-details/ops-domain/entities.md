# 运维管理域 - 领域实体

## Change (变更)
```go
// internal/manager/biz/change/
type Change struct {
    ID          string
    Title       string
    Description string
    Status      ChangeStatus // draft/pending_approval/approved/implementing/completed/cancelled/rejected
    Type        string       // normal/standard/emergency
    Risk        string       // low/medium/high/critical
    Requestor   string
    Approver    string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

## Workflow (工作流)
```go
// internal/manager/biz/workflow/
type Workflow struct {
    ID         string
    Name       string
    Steps      []Step
    Status     string // draft/active/completed/failed
    CreatedAt  time.Time
}

type Step struct {
    ID       string
    Name     string
    Type     string // action/condition/wait
    Config   map[string]interface{}
    Next     []string // 下一步 ID 列表
}
```

## AuditLog (审计日志)
```go
type AuditLog struct {
    ID        string
    EntityID  string
    Action    string
    Actor     string
    Details   string
    Timestamp time.Time
}
```
