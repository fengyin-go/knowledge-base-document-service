# 知识库系统（knowledge-base）

一个纯 Go 标准库（`net/http`，零第三方依赖）实现的企业知识库后端服务。采用 `cmd/internal/pkg` 标准工程分层，内置鉴权/限流中间件，可编译、可测试、可运行。

## 功能特性

- 作者与目录管理：目录支持父子层级，删除前校验子目录/文档引用。
- 文档管理：文档状态机 `draft → published ⇄ archived`，标签自动归一化。
- 版本控制：每次创建/更新自动建立历史版本，支持回滚到任意版本。
- 全文搜索：按标题/内容关键词搜索已发布文档，按浏览量排序。
- 评论与收藏：文档评论、用户收藏。
- 目录树：按父子关系构建完整目录树（含每目录文档数）。
- 批量操作：批量归档、批量删除文档。
- 鉴权中间件：Bearer Token 校验（`AUTH_TOKEN`，为空则放行）。
- 限流中间件：按客户端 IP 固定窗口限流（`RATE_LIMIT`）。
- 请求日志、panic 恢复中间件。
- 多维度统计：文档状态/目录/作者分布、标签统计、热门文档、按月趋势、总览。
- 导入导出：文档 JSON 导入与导出。

## 目录结构

```
origin/
├── cmd/server/main.go
├── internal/
│   ├── app/          # 依赖装配
│   ├── config/       # 环境变量配置
│   ├── middleware/   # 鉴权/限流中间件
│   ├── model/        # 领域模型 + 校验 + 状态机
│   ├── store/        # 数据访问接口 + 内存实现
│   ├── service/      # 业务逻辑（版本/搜索/树/统计）
│   └── handler/      # HTTP 路由 + 处理器
└── pkg/
    ├── httpx/        # 统一响应/分页/JSON
    ├── idgen/        # ID 与短码生成
    └── logger/       # 分级日志
```

## 运行

```bash
cd origin
go run ./cmd/server
```

环境变量：

| 变量 | 默认 | 说明 |
|------|------|------|
| `PORT` | 8080 | 监听端口 |
| `ADDR` | `:8080` | 监听地址（优先于 PORT） |
| `MAX_PAGE_SIZE` | 100 | 分页最大每页条数 |
| `AUTH_TOKEN` | 空 | 鉴权令牌（为空则放行） |
| `RATE_LIMIT` | 0 | 每客户端每分钟最大请求数（0 不限流） |
| `LOG_LEVEL` | info | debug / info / warn / error |

## API 一览

| 方法 | 路径 | 说明 |
|------|------|------|
| POST/GET/PUT/DELETE | `/api/authors` | 作者 CRUD |
| POST/GET/PUT/DELETE | `/api/directories` | 目录 CRUD |
| GET | `/api/directories/tree` | 目录树 |
| POST/GET/PUT/DELETE | `/api/documents` | 文档 CRUD |
| POST | `/api/documents/{id}/publish` | 发布 |
| POST | `/api/documents/{id}/archive` | 归档 |
| GET | `/api/documents/{id}/versions` | 版本列表 |
| POST | `/api/documents/{id}/rollback` | 回滚版本 |
| POST | `/api/documents/batch-archive` | 批量归档 |
| POST | `/api/documents/batch-delete` | 批量删除 |
| GET | `/api/search/documents?q=` | 全文搜索 |
| POST/GET/DELETE | `/api/tags` | 标签管理 |
| POST/GET/DELETE | `/api/comments` | 评论管理 |
| POST/GET/DELETE | `/api/favorites` | 收藏管理 |
| GET | `/api/stats/documents` | 文档统计 |
| GET | `/api/stats/tags` | 标签统计 |
| GET | `/api/stats/authors` | 作者统计 |
| GET | `/api/stats/popular` | 热门文档 |
| GET | `/api/stats/monthly` | 按月趋势 |
| GET | `/api/stats/overview` | 总览 |
| GET | `/api/export/documents` | 导出文档 |
| POST | `/api/import/documents` | 导入文档 |

## 业务闭环示例

```bash
# 1. 建作者 + 目录
curl -s -X POST localhost:8080/api/authors -d '{"name":"张三","email":"z@x.com"}'
curl -s -X POST localhost:8080/api/directories -d '{"name":"技术文档"}'

# 2. 创建文档（带标签）
curl -s -X POST localhost:8080/api/documents -d '{"title":"Go 入门","content":"Go 是编程语言","directory_id":"<dir>","author_id":"<author>","tags":["Go"]}'

# 3. 发布 + 更新（自动建版本）
curl -s -X POST localhost:8080/api/documents/<doc>/publish
curl -s -X PUT localhost:8080/api/documents/<doc> -d '{"content":"Go 是编译型语言"}'

# 4. 查看版本 + 回滚
curl -s localhost:8080/api/documents/<doc>/versions
curl -s -X POST localhost:8080/api/documents/<doc>/rollback -d '{"version_id":"<version_id>"}'

# 5. 搜索
curl -s "localhost:8080/api/search/documents?q=编程"
```

## 测试

```bash
go test ./...
```
