# 基础设施域 - 内部接口

## LLM Client
```go
type LLMClient interface {
    Chat(ctx context.Context, messages []ChatMessage) (string, error)
    StreamChat(ctx context.Context, messages []ChatMessage) (<-chan string, error)
}
```

## RuleEngine
```go
type RuleEngine interface {
    LoadRules(data []byte, format string) error
    Evaluate(ctx context.Context, input string) []RuleMatch
}
```

## AuthManager
```go
type AuthManager interface {
    GenerateToken(claims *JWTClaims) (string, error)
    ValidateToken(token string) (*JWTClaims, error)
}
```
