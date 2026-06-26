# AIOps域 - 服务模块

## ToolsRegistry (工具注册中心)
- **职责**: 注册、发现、执行 26+ 运维工具
- **入口**: `internal/manager/biz/aiops/tools/`
- **核心方法**: Register(), Get(), Execute()
- **内置工具**: promql_query, logql_query, traceql_query, bash_exec, topology_query, knowledge_search, alert_analysis

## AgentCoordinator (Agent 协调器)
- **职责**: Coordinator+Specialist 模式，路由用户请求到合适的 Agent
- **入口**: `internal/manager/biz/aiops/`
- **核心方法**: Route(), Execute()

## CopilotService (Copilot 服务)
- **职责**: 对话式运维助手，维护上下文，调用工具
- **入口**: `internal/manager/biz/aiops/`
- **核心方法**: Chat(), StreamChat()
