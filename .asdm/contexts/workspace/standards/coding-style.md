# 代码风格规范

## 1. Go 规范

### 命名规范
- **包名**: 小写单词，不用下划线 (`httpserver`, `dbx`)
- **导出函数/类型**: PascalCase (`NewUsecase`, `ChangeHandler`)
- **非导出函数/类型**: camelCase (`memRepo`, `buildQuery`)
- **常量**: PascalCase 或 UPPER_SNAKE_CASE
- **文件名**: 小写下划线 (`usecase.go`, `http_handler.go`)

### 包结构 (DDD 分层)
```go
// internal/{bc}/model/    - 领域模型、值对象
// internal/{bc}/biz/      - 业务逻辑 (usecase + repo interface)
// internal/{bc}/server/   - HTTP handler (chi routes)
// internal/{bc}/service/  - 服务层 (跨域协调)
// internal/{bc}/data/     - 数据访问实现
// internal/pkg/           - 共享基础设施
```

### 依赖注入模式
```go
// Usecase 通过构造函数注入 Repo 接口
func NewUsecase(repo ChangeRepo, audit AuditRepo, logger *slog.Logger) *Usecase {
    return &Usecase{repo: repo, audit: audit, logger: logger}
}
```

## 2. HTTP Handler 规范

```go
// Handler 持有 Usecase 引用
type Handler struct {
    uc *biz.Usecase
}

// Register 注册 chi 路由
func (h *Handler) Register(r chi.Router) {
    r.Route("/change", func(r chi.Router) {
        r.Post("/", h.Create)
        r.Get("/", h.List)
        r.Get("/{id}", h.Get)
    })
}
```

## 3. Repository 接口规范

```go
// 接口定义在 biz 层
type ChangeRepo interface {
    Create(ctx context.Context, ch *model.Change) error
    GetByID(ctx context.Context, id string) (*model.Change, error)
    List(ctx context.Context) ([]*model.Change, error)
    Update(ctx context.Context, ch *model.Change) error
}

// 实现在 data 层或使用内存实现
type memChangeRepo struct {
    mu   sync.RWMutex
    data map[string]*model.Change
}
```

## 4. 测试规范

```go
// 表驱动测试
func TestUsecase_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   *model.Change
        wantErr bool
    }{
        {"valid", &model.Change{...}, false},
        {"invalid", nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

---

*本文件由 Context Builder v0.3 生成*
