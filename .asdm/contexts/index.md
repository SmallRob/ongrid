# 项目顶层索引

> **重要**：本文件是 AI 启动时必读的第一个文件，用于建立对项目的全局认知。
> **大小限制**：< 2KB（必须保持精简）

## 1. 项目概览

**项目名称**：Ongrid
**一句话描述**：云边协同 AIOps 平台，通过 AI Agent 理解基础设施、定位根因并自动修复，支持 Slack/Telegram 等 IM 双向交互。
**技术栈定位**：Go + DDD + chi router + MySQL/SQLite + Frontier 隧道

## 2. 技术栈

| 类别 | 技术 |
|------|------|
| 主要语言 | Go 1.25 |
| Web 框架 | chi v5 + net/http |
| 数据库 | MySQL / SQLite (GORM) |
| AI/LLM | Eino 框架 + OpenAI/Anthropic/GLM/DeepSeek/Gemini |
| 向量数据库 | Qdrant (fastembed-go) |
| 边缘通信 | Frontier (geminio SDK) |
| 可观测性 | Prometheus + Loki + Tempo + Grafana |
| 构建工具 | Go modules + Makefile |

## 3. 领域划分

| 领域 | 职责 | 入口 |
|------|------|------|
| [IAM域](domains/iam-domain.md) | 用户认证、组织管理、RBAC权限 | `internal/iam/` |
| [AIOps域](domains/aiops-domain.md) | AI运维智能体、工具集、Agent协调 | `internal/manager/biz/aiops/` |
| [告警域](domains/alert-domain.md) | 告警规则引擎、评估、降噪、关联、自动调查 | `internal/manager/biz/alert/` |
| [可观测性域](domains/observability-domain.md) | 指标采集/查询、日志查询、链路追踪 | `internal/manager/biz/metric/` |
| [边缘域](domains/edge-domain.md) | 边缘设备管理、插件系统、命令策略 | `internal/edgeagent/` |
| [拓扑域](domains/topology-domain.md) | 服务拓扑、设备管理 | `internal/manager/biz/topology/` |
| [知识库域](domains/knowledge-domain.md) | RAG知识管理、代码浏览、向量搜索 | `internal/manager/biz/knowledge/` |
| [运维管理域](domains/ops-domain.md) | 变更管理、工作流引擎、报告生成 | `internal/manager/biz/change/` |
| [基础设施域](domains/infrastructure-domain.md) | LLM客户端、嵌入模型、认证、配置、通知 | `internal/pkg/` |

## 4. 构建指南

```bash
# 构建云端服务
go build -o ongrid ./cmd/ongrid

# 构建边缘代理
go build -o ongrid-edge ./cmd/ongrid-edge

# 运行测试
go test ./...

# Docker 部署
docker-compose up -d
```

## 5. 导航指引

- **理解项目**：阅读本文件（L1）
- **修改某领域**：阅读对应领域索引（L2）
- **执行具体任务**：按需阅读详细内容（L3）

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
