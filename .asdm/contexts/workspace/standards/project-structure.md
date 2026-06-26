# 项目结构规范

## 1. 目录组织 (DDD + Bounded Context)

```
ongrid/
├── cmd/
│   ├── ongrid/          # 云端服务入口 (main.go)
│   └── ongrid-edge/     # 边缘代理入口 (main.go)
├── internal/
│   ├── iam/             # IAM 有界上下文
│   │   ├── biz/         # 业务逻辑
│   │   │   ├── authz/   # 授权 (Casbin)
│   │   │   ├── user/    # 用户管理
│   │   │   ├── org/     # 组织管理
│   │   │   └── membership/ # 成员关系
│   │   ├── model/       # 领域模型
│   │   ├── server/      # HTTP handler
│   │   ├── service/     # 服务层
│   │   └── data/        # 数据访问实现
│   ├── manager/         # Manager 有界上下文
│   │   ├── biz/         # 业务逻辑
│   │   │   ├── aiops/   # AIOps (tools/)
│   │   │   ├── alert/   # 告警引擎
│   │   │   ├── metric/  # 指标处理
│   │   │   ├── topology/# 拓扑管理
│   │   │   ├── knowledge/# 知识库
│   │   │   ├── change/  # 变更管理
│   │   │   ├── workflow/# 工作流
│   │   │   ├── edge/    # 边缘管理
│   │   │   ├── report/  # 报告生成
│   │   │   └── ...      # 其他子域
│   │   ├── model/       # 领域模型
│   │   ├── server/      # HTTP handler
│   │   ├── service/     # 服务层
│   │   └── data/        # 数据访问实现
│   ├── edgeagent/       # 边缘代理有界上下文
│   │   ├── biz/         # 业务逻辑
│   │   ├── cmdpolicy/   # 命令策略引擎
│   │   ├── plugins/     # 插件系统
│   │   ├── bash/        # Bash 处理器
│   │   ├── host_files/  # 文件处理器
│   │   ├── collector/   # 采集器
│   │   └── model/       # 领域模型
│   └── pkg/             # 共享基础设施
│       ├── llm/         # LLM 客户端
│       ├── auth/        # JWT 认证
│       ├── config/      # 配置
│       ├── dbx/         # 数据库
│       ├── embedding/   # 嵌入
│       ├── notify/      # 通知
│       ├── authzmw/     # 鉴权中间件
│       ├── tracequery/  # 追踪查询
│       └── ...          # 其他
├── api/                 # Protobuf 定义
├── deploy/              # 部署配置
├── agents/              # Agent 角色定义
├── docs/                # 文档
└── Makefile
```

## 2. 文件命名

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| 入口 | main.go | `cmd/ongrid/main.go` |
| 业务逻辑 | usecase.go | `biz/change/usecase.go` |
| 仓库接口 | repo.go | `biz/change/repo.go` |
| 领域模型 | model.go | `model/change.go` |
| HTTP处理 | http.go / {name}.go | `server/change.go` |
| 测试 | {name}_test.go | `usecase_test.go` |
| 配置 | config.go | `pkg/config/config.go` |

## 3. 模块职责边界

- **cmd**: 程序入口，组装依赖
- **biz**: 业务逻辑，定义 Repo 接口，不依赖具体实现
- **model**: 领域模型和值对象
- **server**: HTTP handler，调用 biz/usecase
- **service**: 跨域协调服务
- **data**: Repo 接口的具体实现 (SQL/Memory)
- **pkg**: 可复用的基础设施组件

## 4. 新增子域清单

添加新的业务子域时需要创建的目录和文件：
1. `internal/manager/biz/{domain}/` - 业务逻辑
2. `internal/manager/biz/{domain}/repo.go` - 仓库接口
3. `internal/manager/biz/{domain}/usecase.go` - 业务用例
4. `internal/manager/model/{domain}.go` - 领域模型
5. `internal/manager/server/{domain}/http.go` - HTTP handler
6. `cmd/ongrid/main.go` - 注册路由和依赖注入

---

*本文件由 Context Builder v0.3 生成*
