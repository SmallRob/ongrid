# 部署配置

## 1. 环境变量

主要环境变量配置（参考 `deploy/` 目录）：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| ONGRID_DB_DSN | 数据库连接串 | sqlite:ongrid.db |
| ONGRID_JWT_SECRET | JWT 签名密钥 | - |
| ONGRID_FRONTIER_ADDR | Frontier 服务地址 | - |
| ONGRID_PROM_URL | Prometheus 地址 | http://localhost:9090 |
| ONGRID_LOKI_URL | Loki 地址 | http://localhost:3100 |
| ONGRID_TEMPO_URL | Tempo 地址 | http://localhost:3200 |
| ONGRID_QDRANT_URL | Qdrant 地址 | http://localhost:6333 |
| ONGRID_LLM_PROVIDER | LLM Provider 名称 | openai |
| ONGRID_LLM_API_KEY | LLM API Key | - |

## 2. Docker 部署

```bash
# 一键部署完整栈 (含 Prometheus/Grafana/Loki/Tempo/Qdrant)
cd deploy && ./install.sh

# 或使用 docker-compose
docker-compose up -d
```

## 3. Systemd 部署

```bash
# 安装为系统服务
cd deploy/install/systemd && sudo ./install-systemd.sh

# 服务列表
# - ongrid.service         云端服务
# - ongrid-frontier.service Frontier 隧道
# - prometheus.service     Prometheus
# - loki.service           Loki
# - tempo.service          Tempo
```

## 4. 边缘部署

```bash
# 打包边缘 bundle
deploy/install/edge/build-edge-bundle.sh

# 在目标机器安装
sudo ./install-edge.sh
```

## 5. 数据目录

- 数据库文件: 取决于配置 (SQLite 或 MySQL)
- 向量数据库: Qdrant (远程或本地)
- 日志: Loki + 本地 stdout
- 追踪: Tempo
- 配置: `deploy/install/frontier.yaml`

---

*本文件由 Context Builder v0.3 生成*
