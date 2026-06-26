# 知识库域 - RAG知识管理

> **大小限制**：< 1KB（必须保持精简）
> **作用**：域级入口，快速了解某个业务领域的组成

## 核心职责

RAG 知识库管理、文档导入、代码浏览、向量检索、嵌入生成。

## 子模块

| 子模块 | 职责 | 入口 |
|--------|------|------|
| 知识库 (knowledge) | 知识 CRUD/检索 | `internal/manager/biz/knowledge/` |
| 嵌入 (embedding) | 本地嵌入模型 | `internal/pkg/embedding/` |
| 向量存储 | Qdrant 向量操作 | `internal/pkg/qdrantx/` |
| 代码浏览 | 代码库索引 | `internal/manager/biz/knowledge/codebrowser/` |

## 关键接口

- `POST /api/v1/knowledge` - 创建知识条目
- `GET /api/v1/knowledge/search` - 向量检索
- `POST /api/v1/knowledge/import` - 批量导入

## 关键依赖

- Qdrant - 向量数据库
- fastembed-go - 本地嵌入模型

---

*本文件由 Context Builder v0.3 生成，保持精简以确保 AI 高效加载*
