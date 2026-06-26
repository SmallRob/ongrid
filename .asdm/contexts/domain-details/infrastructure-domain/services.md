# 基础设施域 - 服务模块

## LLMClient (LLM 客户端)
- **职责**: 多 Provider 路由、自动故障转移、重试
- **入口**: `internal/pkg/llm/`
- **支持**: OpenAI, Anthropic, GLM(Zhipu), DeepSeek, Gemini
- **核心方法**: Chat(), StreamChat(), ListProviders()

## EmbeddingService (嵌入服务)
- **职责**: 本地向量嵌入 (fastembed-go)
- **入口**: `internal/pkg/embedding/`
- **核心方法**: Embed(), EmbedBatch()

## AuthManager (认证管理)
- **职责**: JWT 令牌签发/验证
- **入口**: `internal/pkg/auth/`
- **核心方法**: GenerateToken(), ValidateToken()

## RuleEngine (规则引擎)
- **职责**: 关键词+正则规则匹配、YAML/JSON 规则加载
- **入口**: `internal/pkg/ruleengine/`
- **核心方法**: LoadRules(), Evaluate(), Match()

## NotifyService (通知服务)
- **职责**: 多渠道通知 (Slack/Telegram/飞书/钉钉)
- **入口**: `internal/pkg/notify/`
- **核心方法**: Send(), SendBatch()
