# 告警域 - 告警规则引擎

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

告警规则管理、告警评估、降噪、关联分析、自动调查、告警事件处理。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 规则 (rule) | 告警规则 CRUD | `internal/manager/biz/alert/rule/` |
| 评估 (eval) | 规则评估引擎 | `internal/manager/biz/alert/eval/` |
| 降噪 (dedup) | 告警去重/降噪 | `internal/manager/biz/alert/dedup/` |
| 关联 (corr) | 告警关联分析 | `internal/manager/biz/alert/corr/` |
| 调查 (investigate) | 自动根因调查 | `internal/manager/biz/alert/investigate/` |

## 关键接口

- `POST /api/v1/alerts/rules` - 创建规则
- `GET /api/v1/alerts/events` - 告警事件列表
- `POST /api/v1/alerts/evaluate` - 手动触发评估

## 关键依赖

- Prometheus - 指标查询
- 规则引擎 (`internal/pkg/ruleengine/`) - 规则匹配

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
