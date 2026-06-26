# AIOps域 - 领域实体

## Tool (工具)
```go
// internal/manager/biz/aiops/tools/
type Tool struct {
    Name        string   // 工具名称
    Description string   // 工具描述
    Parameters  []Param  // 参数定义
    Handler     func(ctx context.Context, params map[string]interface{}) (string, error)
}
```

## Agent (智能体)
```go
type Agent struct {
    Name         string   // Agent 名称
    Role         string   // 角色描述
    SystemPrompt string   // 系统提示词
    Tools        []string // 可用工具列表
    Specialties  []string // 专长领域
}
```

## ChatMessage (对话消息)
```go
type ChatMessage struct {
    Role    string // user/assistant/system
    Content string
}
```
