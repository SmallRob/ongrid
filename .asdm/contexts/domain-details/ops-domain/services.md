# 运维管理域 - 服务模块

## ChangeManager (变更管理器)
- **职责**: 变更生命周期管理 (创建/审批/实施/完成/取消)
- **入口**: `internal/manager/biz/change/`
- **核心方法**: Create(), Approve(), Reject(), Implement(), Complete(), Cancel()

## WorkflowEngine (工作流引擎)
- **职责**: 工作流定义、执行、状态管理
- **入口**: `internal/manager/biz/workflow/`
- **核心方法**: Create(), Start(), Step(), Complete()

## ReportGenerator (报告生成器)
- **职责**: 运维报告自动生成
- **入口**: `internal/manager/biz/report/`
- **核心方法**: Generate(), List(), Get()

## IMBridge (IM 桥接)
- **职责**: Slack/Telegram/飞书/钉钉 消息双向交互
- **入口**: `internal/manager/biz/imbridge/`
- **核心方法**: Send(), Receive(), RegisterHandler()
