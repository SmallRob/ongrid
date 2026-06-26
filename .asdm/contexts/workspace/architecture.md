# 项目架构概览

> 📊 可视化架构图：[SVG](../../../docs/architecture/ongrid-architecture-v1.svg) | [PNG](../../../docs/architecture/ongrid-architecture-v1.png)

## 1. 整体架构

```
┌──────────────────────────────────────────────────────────────┐
│                      Ongrid Cloud Side                        │
├──────────────────────────────────────────────────────────────┤
│  IM Channels (Slack / Telegram / Larksuite / DingTalk)       │
├──────────────────────────────────────────────────────────────┤
│  HTTP API (chi router)                                       │
│  ├── /api/v1/iam/*       用户认证、组织管理                   │
│  ├── /api/v1/aiops/*     AI Agent 对话、工具调用              │
│  ├── /api/v1/alerts/*    告警规则、告警事件                   │
│  ├── /api/v1/metrics/*   指标查询                             │
│  ├── /api/v1/topology/*  拓扑管理                             │
│  ├── /api/v1/edges/*     边缘设备管理                         │
│  ├── /api/v1/knowledge/* 知识库                               │
│  ├── /api/v1/change/*    变更管理                             │
│  ├── /api/v1/workflow/*  工作流                               │
│  └── /metrics            Prometheus 指标                      │
├──────────────────────────────────────────────────────────────┤
│  Business Layer (DDD Bounded Contexts)                       │
│  ├── iam/          身份与访问管理                             │
│  ├── manager/      核心业务管理                               │
│  │   ├── aiops/    AI 运维智能体                              │
│  │   ├── alert/    告警引擎                                   │
│  │   ├── metric/   指标处理                                   │
│  │   ├── topology/ 拓扑管理                                   │
│  │   ├── knowledge/知识库                                     │
│  │   ├── change/   变更管理                                   │
│  │   ├── workflow/ 工作流引擎                                 │
│  │   └── ...      其他子域                                    │
│  └── shared (pkg/) 共享基础设施                               │
├──────────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                        │
│  ├── LLM Client (Eino + 多Provider路由)                      │
│  ├── Embedding (fastembed-go 本地模型)                       │
│  ├── Vector Store (Qdrant)                                   │
│  ├── Database (MySQL/SQLite via GORM)                        │
│  ├── Observability (Prometheus + Loki + Tempo)               │
│  └── Frontier Tunnel (geminio SDK)                           │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    Ongrid Edge Side                           │
├──────────────────────────────────────────────────────────────┤
│  Edge Agent (ongrid-edge)                                    │
│  ├── Plugin System (host/metrics/logs/traces/custom/db)      │
│  ├── Command Policy Engine (沙箱 + 分级管控)                 │
│  ├── Collector (CPU/MEM/NET/进程)                            │
│  ├── Bash Handler (远程命令执行)                             │
│  ├── Host Files (远程文件操作)                               │
│  ├── Service Handler (服务管理)                              │
│  └── WebShell Handler (WebSocket终端)                        │
└──────────────────────────────────────────────────────────────┘
```

## 2. 技术栈详情

### 云端 (cmd/ongrid)
- **语言**: Go 1.25
- **路由**: chi v5 + net/http
- **数据库**: GORM + MySQL/SQLite
- **AI框架**: Eino (cloudwego/eino)
- **向量搜索**: Qdrant + fastembed-go
- **认证**: JWT (golang-jwt) + Casbin RBAC
- **隧道**: Frontier (singchia/geminio)
- **可观测**: Prometheus client_golang + OpenTelemetry

### 边缘 (cmd/ongrid-edge)
- **语言**: Go 1.25
- **通信**: Frontier SDK (geminio)
- **采集**: gopsutil (CPU/MEM/NET)
- **安全**: 命令策略引擎 + Bash 沙箱

## 3. 目录结构

```
ongrid/
├── cmd/
│   ├── ongrid/          # 云端服务入口
│   └── ongrid-edge/     # 边缘代理入口
├── internal/
│   ├── iam/             # IAM 有界上下文
│   │   ├── biz/         # 业务逻辑 (authz/membership/org/user)
│   │   ├── model/       # 领域模型
│   │   ├── server/      # HTTP handler
│   │   └── service/     # 服务层
│   ├── manager/         # Manager 有界上下文
│   │   ├── biz/         # 业务逻辑 (aiops/alert/metric/topology/...)
│   │   ├── model/       # 领域模型
│   │   ├── server/      # HTTP handler
│   │   └── service/     # 服务层
│   ├── edgeagent/       # 边缘代理有界上下文
│   │   ├── biz/         # 业务逻辑 (agent/collector/upgrade)
│   │   ├── cmdpolicy/   # 命令策略引擎
│   │   ├── plugins/     # 插件系统
│   │   └── ...          # 各处理器
│   └── pkg/             # 共享基础设施包
│       ├── llm/         # LLM 客户端 + 多Provider路由
│       ├── auth/        # JWT 认证
│       ├── config/      # 配置管理
│       ├── dbx/         # 数据库封装
│       ├── embedding/   # 嵌入模型
│       ├── notify/      # 通知
│       ├── authzmw/     # 鉴权中间件
│       ├── tracequery/  # 追踪查询
│       └── ...          # 其他共享包
├── api/                 # Protobuf API 定义
├── deploy/              # 部署配置 (Docker/systemd/脚本)
├── agents/              # Agent 角色定义 (Markdown)
└── docs/                # 文档
```

## 4. 关键设计决策

### 4.1 DDD + 有界上下文
- 3 个核心有界上下文: IAM、Manager、EdgeAgent
- 每个有界上下文内部遵循 biz/model/server 分层
- 共享基础设施通过 `internal/pkg/` 复用

### 4.2 云边协同架构
- 边缘零入站端口：边缘主动拨出连接 Frontier
- Manager 通过 Frontier 隧道反向调用边缘工具
- 边缘插件系统支持热插拔

### 4.3 AI Agent 架构
- Coordinator + Specialist 模式
- 26+ 内置工具 (PromQL/LogQL/TraceQL/Bash/拓扑/知识库)
- 多 LLM Provider 动态路由 + 自动故障转移

---

*本文件由 Context Builder v0.3 生成*
