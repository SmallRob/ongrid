# IAM域 - 身份与访问管理

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

用户认证、组织管理、RBAC 权限控制、成员关系管理。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 认证 (auth) | JWT 令牌签发/验证 | `internal/iam/biz/auth/` |
| 用户 (user) | 用户 CRUD | `internal/iam/biz/user/` |
| 组织 (org) | 组织管理 | `internal/iam/biz/org/` |
| 成员 (membership) | 组织-用户关联 | `internal/iam/biz/membership/` |
| 授权 (authz) | Casbin RBAC 策略 | `internal/iam/biz/authz/` |

## 关键接口

- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/register` - 用户注册
- `GET /api/v1/users/me` - 当前用户信息
- `GET /api/v1/orgs` - 组织列表

## 关键依赖

- JWT (golang-jwt) - 令牌签发
- Casbin - RBAC 策略引擎
- GORM - 数据持久化

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
