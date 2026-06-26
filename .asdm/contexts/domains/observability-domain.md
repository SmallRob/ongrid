# 可观测性域 - 指标/日志/链路

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

指标采集与查询、日志查询、链路追踪查询、Prometheus Remote Write 代理。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 指标 (metric) | 指标查询/聚合 | `internal/manager/biz/metric/` |
| 日志 (logquery) | Loki 日志查询 | `internal/manager/biz/metric/logquery/` |
| 链路 (trace) | Tempo 追踪查询 | `internal/manager/biz/metric/trace/` |
| Prometheus 写入 | Remote Write 代理 | `internal/manager/biz/promwrite/` |
| Grafana 代理 | Grafana 代理接口 | `internal/manager/biz/grafana/` |
| 监控 (monitor) | 系统监控 | `internal/manager/biz/monitor/` |

## 关键接口

- `GET /api/v1/metrics/query` - PromQL 查询
- `GET /api/v1/logs/query` - LogQL 查询
- `GET /api/v1/traces/query` - TraceQL 查询
- `POST /api/v1/prom/write` - Prometheus 写入

## 关键依赖

- Prometheus client_golang - 指标客户端
- Loki HTTP API - 日志查询
- Tempo HTTP API - 追踪查询

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
