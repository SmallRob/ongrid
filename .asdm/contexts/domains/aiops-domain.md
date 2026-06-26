# AIOps域 - AI运维智能体

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

AI 运维智能体编排、工具注册与调用、Copilot 对话、Agent 角色管理。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 工具注册 (tools) | 26+ 内置运维工具 | `internal/manager/biz/aiops/tools/` |
| Agent 编排 | Coordinator+Specialist 模式 | `internal/manager/biz/aiops/` |
| Copilot | 对话式运维助手 | `internal/manager/biz/aiops/` |
| Agent 角色 | Markdown 角色定义 | `agents/` |

## 关键接口

- `POST /api/v1/aiops/chat` - AI 对话
- `GET /api/v1/aiops/tools` - 工具列表
- `POST /api/v1/aiops/execute` - 执行工具

## 工具集

PromQL 查询、LogQL 查询、TraceQL 查询、Bash 执行、拓扑查询、知识库检索、告警分析、指标查询等。

## 关键依赖

- Eino 框架 (cloudwego/eino) - AI 编排
- LLM Client (`internal/pkg/llm/`) - 多 Provider 路由
- 向量搜索 (Qdrant) - 知识库检索

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
