# 边缘域 - API 端点

## 边缘设备管理
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/edges | 边缘设备列表 |
| GET | /api/v1/edges/{id} | 设备详情 |
| POST | /api/v1/edges/{id}/execute | 远程执行命令 |
| DELETE | /api/v1/edges/{id} | 移除设备 |

## 远程操作
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/edges/{id}/shell | WebShell WebSocket |
| GET | /api/v1/edges/{id}/files | 远程文件浏览 |
| POST | /api/v1/edges/{id}/upgrade | 触发升级 |
