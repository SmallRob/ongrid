# 运维管理域 - 变更/工作流/报告

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

ITSM 变更管理、工作流引擎、报告生成、运维市场。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 变更 (change) | 变更生命周期管理 | `internal/manager/biz/change/` |
| 工作流 (workflow) | 工作流定义/执行 | `internal/manager/biz/workflow/` |
| 报告 (report) | 报告生成 | `internal/manager/biz/report/` |
| 市场 (marketplace) | 插件/模板市场 | `internal/manager/biz/marketplace/` |
| WebShell | 远程终端 | `internal/manager/biz/webshell/` |
| IM 桥接 | Slack/Telegram/飞书/钉钉 | `internal/manager/biz/imbridge/` |

## 关键接口

- `POST /api/v1/change` - 创建变更
- `POST /api/v1/change/{id}/approve` - 审批变更
- `GET /api/v1/workflow` - 工作流列表
- `POST /api/v1/report` - 生成报告

## 关键依赖

- 规则引擎 (`internal/pkg/ruleengine/`) - 工作流规则
- IM Bridge - 多渠道通知

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
