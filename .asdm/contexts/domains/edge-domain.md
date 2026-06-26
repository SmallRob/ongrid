# 边缘域 - 边缘设备管理

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

边缘设备注册/管理、插件系统、命令策略引擎、远程命令执行、采集器。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 边缘管理 (edge) | 设备注册/心跳 | `internal/manager/biz/edge/` |
| 插件系统 (plugins) | 热插拔插件 | `internal/edgeagent/plugins/` |
| 命令策略 (cmdpolicy) | 5级分类+沙箱 | `internal/edgeagent/cmdpolicy/` |
| Bash 处理 (bash) | 远程命令执行 | `internal/edgeagent/bash/` |
| 文件操作 (host_files) | 远程文件管理 | `internal/edgeagent/host_files/` |
| 采集器 (collector) | 系统指标采集 | `internal/edgeagent/collector/` |
| WebShell | WebSocket 终端 | `internal/edgeagent/` |

## 关键接口

- `GET /api/v1/edges` - 边缘设备列表
- `POST /api/v1/edges/{id}/execute` - 远程执行
- `GET /api/v1/edges/{id}/shell` - WebShell WebSocket

## 关键依赖

- Frontier SDK (geminio) - 云边隧道
- gopsutil - 系统信息采集
- 命令策略引擎 - 安全管控

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
