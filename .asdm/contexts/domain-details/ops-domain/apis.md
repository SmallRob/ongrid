# 运维管理域 - API 端点

## 变更管理
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/change | 创建变更 |
| GET | /api/v1/change | 变更列表 |
| GET | /api/v1/change/{id} | 变更详情 |
| POST | /api/v1/change/{id}/approve | 审批通过 |
| POST | /api/v1/change/{id}/reject | 审批拒绝 |
| POST | /api/v1/change/{id}/implement | 开始实施 |
| POST | /api/v1/change/{id}/complete | 完成变更 |

## 工作流
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/workflow | 创建工作流 |
| GET | /api/v1/workflow | 工作流列表 |
| POST | /api/v1/workflow/{id}/start | 启动执行 |
| GET | /api/v1/workflow/{id}/status | 执行状态 |

## 报告
| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/report | 生成报告 |
| GET | /api/v1/report | 报告列表 |
| GET | /api/v1/report/{id} | 报告详情 |
