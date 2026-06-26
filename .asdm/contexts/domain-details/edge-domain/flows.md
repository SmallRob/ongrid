# 边缘域 - 核心业务流程

## 云边命令执行流程
1. Manager 收到执行请求
2. CommandPolicyEngine.Classify() 分类命令安全级别
3. 命令策略校验 (路径沙箱/网络白名单)
4. 通过 Frontier 隧道发送到边缘
5. 边缘 Agent 接收命令
6. 本地 CommandPolicy 再次校验
7. 执行命令并收集结果
8. 通过 Frontier 返回结果

## 插件加载流程
1. 边缘 Agent 启动
2. 扫描插件目录
3. 逐个加载插件配置
4. 初始化插件实例
5. 注册插件处理器
6. 开始采集/处理数据
