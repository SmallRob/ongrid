# 拓扑域 - 服务拓扑管理

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

服务拓扑建模、设备管理、拓扑发现、关系维护。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 拓扑 (topology) | 服务/资源拓扑 | `internal/manager/biz/topology/` |
| 设备 (device) | 设备管理 | `internal/manager/biz/device/` |
| 拓扑发现 | 自动发现服务关系 | `internal/manager/biz/topology/` |

## 关键接口

- `GET /api/v1/topology` - 拓扑图
- `POST /api/v1/topology/nodes` - 创建节点
- `GET /api/v1/devices` - 设备列表

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
