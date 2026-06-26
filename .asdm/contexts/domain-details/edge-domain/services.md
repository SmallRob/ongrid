# 边缘域 - 服务模块

## EdgeManager (边缘管理器)
- **职责**: 边缘设备注册、心跳管理、状态监控
- **入口**: `internal/manager/biz/edge/`
- **核心方法**: Register(), Heartbeat(), List()

## PluginSystem (插件系统)
- **职责**: 插件加载、热插拔、生命周期管理
- **入口**: `internal/edgeagent/plugins/`
- **核心方法**: Load(), Unload(), Enable(), Disable()

## CommandPolicyEngine (命令策略引擎)
- **职责**: 5级安全分类、读写分离、路径沙箱、网络白名单
- **入口**: `internal/edgeagent/cmdpolicy/`
- **核心方法**: Classify(), Validate(), Execute()

## Collector (采集器)
- **职责**: 系统指标采集 (CPU/MEM/NET/进程)
- **入口**: `internal/edgeagent/collector/`
- **核心方法**: Collect(), Start(), Stop()
