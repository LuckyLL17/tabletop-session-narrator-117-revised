# 桌游局势与战报工作台

这是一个面向桌游爱好者的 Go 全栈记录工具。它不负责玩某一款具体桌游，而是把一局桌游从准备、开始、暂停、回合推进、行动记录到结束复盘的过程保存成可重放的时间线，并生成一份能解释“哪里发生了转折”的战报。

## 问题背景

桌游现场通常只有零散的口头记忆：谁在第几回合做了什么、什么时候资源发生了变化、哪个动作改变了局面，结束后很快就难以还原。本项目提供一个轻量的个人工作台，让玩家在不打断游戏的情况下记录行动和关键事件，结束后回看完整过程。

## 领域术语

- **桌游档案**：记录一款游戏的玩家人数范围、默认资源和规则变体。
- **对局**：一次具体的游戏过程，状态从“准备中”流转到“进行中”、可短暂“暂停”，最后进入“已结束”。
- **回合**：由当前玩家负责的一段行动窗口，编号在同一对局内唯一。
- **行动事件**：绑定回合和玩家的事实记录，可带有得分变化或资源变化。
- **关键时刻**：由高影响力行动自动产生或手工标记的转折点。
- **战报**：基于对局快照、回合、行动和关键时刻生成的赛后叙事。

## 核心流程

1. 注册或登录个人工作台。
2. 创建桌游档案，设定玩家人数和默认资源。
3. 创建对局并填写玩家席位，可配置规则变体。
4. 开始对局，按当前玩家推进回合，记录行动、得分和资源变化。
5. 用暂停/继续表达现场节奏，使用高影响力行动标记关键转折。
6. 结束对局，后台任务生成战报，前端展示玩家策略线、赢家和复盘问题。
7. 重放时间线或下载 Markdown/JSON 作为个人记录。

## 模块说明

- `internal/domain`：状态机、不变量、领域对象和输入规则。
- `internal/store`：独立的 JSON 快照与追加事件日志持久化，写入采用临时文件替换。
- `internal/service`：认证、桌游档案、对局、回合、行动、资源、时间线、重放、战报、分析、比较、复盘和导出。
- `internal/jobs`：结束对局后的战报任务、轮询、重试状态和优雅停止。
- `internal/httpapi`：真实 HTTP API、中间件和嵌入式前端服务。
- `web`：原生 HTML、CSS、JavaScript 浏览器工作台。

## 启动方式

```bash
GOCACHE=/tmp/tabletop117-gocache go run ./cmd/server -addr 127.0.0.1:8098 -data /tmp/tabletop117.json
```

然后打开 `http://127.0.0.1:8098/`。也可以通过 `-secret` 或 `TABLETOP_SECRET` 替换会话签名密钥。

## API 摘要

- `POST /api/v1/auth/register`、`POST /api/v1/auth/login`、`GET /api/v1/session`
- `GET/POST /api/v1/games`、`POST /api/v1/games/{id}/variants`
- `GET/POST /api/v1/matches`、`GET /api/v1/matches/{id}`
- `POST /api/v1/matches/{id}/start|pause|resume|finish`
- `POST /api/v1/matches/{id}/turns`、`POST /api/v1/matches/{id}/events`
- `GET /api/v1/matches/{id}/timeline`、`GET /api/v1/matches/{id}/replay`
- `GET /api/v1/matches/{id}/report`
- `GET /api/v1/matches/{id}/analysis|suggestions|scorecard|catalog|insights`
- `GET/POST /api/v1/matches/{id}/reflections`
- `GET /api/v1/compare?match_id=...&match_id=...`
- `GET /api/v1/search?q=关键词`
- `GET /api/v1/export/markdown?match_id=...`、`GET /api/v1/export/json?match_id=...`

## 验收方式

```bash
GOCACHE=/tmp/tabletop117-gocache go build ./...
GOCACHE=/tmp/tabletop117-gocache go vet ./...
GOCACHE=/tmp/tabletop117-gocache go test ./...
```

浏览器或 HTTP 客户端应至少验证：注册、创建桌游、创建对局、开始对局、开启回合、写入行动、读取时间线和分析、结束对局、读取战报、评分卡与洞察摘要、保存复盘回答和下载导出文件。项目按技能要求不包含 `*_test.go` 测试文件，`go test` 仅用于确认所有包可编译。

生成日期：2026-08-24。
