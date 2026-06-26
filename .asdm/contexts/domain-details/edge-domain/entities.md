# 边缘域 - 领域实体

## EdgeDevice (边缘设备)
```go
// internal/manager/biz/edge/
type EdgeDevice struct {
    ID       string // 设备 ID
    Name     string // 设备名称
    Address  string // 网络地址
    Status   string // online/offline
    LastSeen time.Time
}
```

## Plugin (插件)
```go
// internal/edgeagent/plugins/
type Plugin struct {
    Name     string                 // 插件名称
    Version  string                 // 版本
    Type     string                 // host/metrics/logs/traces/custom/db
    Config   map[string]interface{} // 配置
    Enabled  bool
}
```

## CommandPolicy (命令策略)
```go
// internal/edgeagent/cmdpolicy/
type CommandPolicy struct {
    Level      SecurityLevel // READ_FS/READ_SYSTEM/MIXED/NETWORK/DENIED
    Command    string
    Arguments  []string
    WorkingDir string
    EnvVars    map[string]string
}
```
