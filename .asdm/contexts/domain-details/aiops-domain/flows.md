# AIOps域 - 核心业务流程

## AI 对话流程
1. 用户发送消息 → CopilotService.Chat()
2. 检查 Token 预算 → 确认配额
3. 加载 Agent 角色 → 读取 SystemPrompt
4. 构建消息上下文 → 历史消息 + 系统提示
5. 调用 LLM → 多 Provider 路由 + 故障转移
6. 解析工具调用 → ToolRegistry.Execute()
7. 返回结果 → 流式/同步

## 工具执行流程
1. LLM 返回 tool_calls
2. 解析工具名和参数
3. ToolRegistry 查找工具
4. 参数校验
5. 执行工具 Handler
6. 返回执行结果
7. 继续对话循环
