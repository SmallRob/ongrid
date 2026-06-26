# AIOps域 - API 端点

## 对话接口
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/aiops/chat | AI 对话（同步） |
| POST | /api/v1/aiops/chat/stream | AI 对话（流式） |

## 工具接口
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/aiops/tools | 获取工具列表 |
| POST | /api/v1/aiops/tools/{name}/execute | 执行指定工具 |

## Agent 接口
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/aiops/agents | Agent 角色列表 |
| POST | /api/v1/aiops/agents/{name}/chat | 与指定 Agent 对话 |
