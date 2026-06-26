# 基础设施域 - 核心实体

## LLMProvider (LLM Provider)
```go
// internal/pkg/llm/
type LLMProvider struct {
    Name     string // openai/anthropic/glm/deepseek/gemini
    APIKey   string
    BaseURL  string
    Priority int    // 优先级 (越小越高)
    Enabled  bool
}
```

## JWTClaims
```go
// internal/pkg/auth/
type JWTClaims struct {
    UserID string
    OrgID  string
    Role   string
    jwt.RegisteredClaims
}
```

## Rule (规则)
```go
// internal/pkg/ruleengine/
type Rule struct {
    ID          string
    Name        string
    Pattern     string // 正则模式
    Keywords    []string
    Severity    string
    Action      string
    Enabled     bool
}
```
