# 基础设施域 - 共享组件

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

LLM 多 Provider 路由、嵌入模型、JWT 认证、数据库封装、配置管理、通知渠道、规则引擎。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| LLM Client | 多 Provider 路由/故障转移 | `internal/pkg/llm/` |
| 嵌入模型 | 本地嵌入 (fastembed-go) | `internal/pkg/embedding/` |
| JWT 认证 | 令牌签发/验证 | `internal/pkg/auth/` |
| 数据库 | GORM 封装 (MySQL/SQLite) | `internal/pkg/dbx/` |
| 配置 | 统一配置管理 | `internal/pkg/config/` |
| 通知 | 多渠道通知 | `internal/pkg/notify/` |
| 规则引擎 | 关键词/正则规则匹配 | `internal/pkg/ruleengine/` |
| 授权 (authz) | Casbin RBAC 策略 | `internal/iam/biz/authz/` |
| 追踪查询 | TraceQL 查询客户端 | `internal/pkg/tracequery/` |
| 鉴权中间件 | RBAC 中间件 | `internal/pkg/authzmw/` |

## 关键依赖

- Eino (cloudwego/eino) - AI 框架
- Casbin - RBAC
- GORM - ORM
- Frontier (singchia/geminio) - 云边通信 (独立部署)

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
