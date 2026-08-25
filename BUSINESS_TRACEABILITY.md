# 业务可达性追踪

本文件逐一说明生产 Go 文件的业务入口、下游调用和关键规则。项目没有测试专用构件，所有列出的文件都从 HTTP、启动任务或服务调用链可达。

| 文件 | 业务场景 | 生产调用链 |
|---|---|---|
| `cmd/server/main.go` | 启动服务、组装依赖、优雅停止 | 进程启动 -> Store/Auth/Services/HTTP/Worker |
| `internal/domain/types.go` | 对局、回合、行动和战报模型 | 全部领域与接口层 |
| `internal/domain/analysis_types.go` | 玩家分析、资源轨迹、对局比较、复盘评分模型 | MatchService 分析/比较/评分卡与 HTTP 分析接口 |
| `internal/domain/insight_types.go` | 战报洞察摘要、玩家卡片和资源卡片模型 | MatchService 洞察接口与浏览器工作台 |
| `internal/domain/enums.go` | 事件类型和状态判断 | MatchService/Event API |
| `internal/domain/errors.go` | 领域错误分类 | Domain -> HTTP 状态映射 |
| `internal/domain/validate.go` | 桌游、席位、回合和行动校验 | GameService/MatchService |
| `internal/domain/transitions.go` | 对局状态机与回合轮转 | MatchService |
| `internal/domain/factory.go` | 创建领域对象和复制资源 | GameService/MatchService |
| `internal/store/store.go` | JSON 快照事务更新 | 所有服务写入 |
| `internal/store/journal.go` | 追加事件日志 | MatchService 创建/状态变更 |
| `internal/store/users.go` | 用户持久化和查找 | AuthService |
| `internal/store/games.go` | 桌游档案与变体 | GameService |
| `internal/store/matches.go` | 对局、席位、回合、行动、关键时刻 | MatchService |
| `internal/store/reports.go` | 战报和复盘持久化 | ReportService |
| `internal/store/jobs.go` | 任务领取和状态推进 | JobService/Worker |
| `internal/store/queries.go` | 历史搜索和数据规模 | SearchService/Health |
| `internal/store/analysis_queries.go` | 按席位、回合、时间和标签查询事实记录 | 分析、比较、复盘和行动目录服务 |
| `pkg/ids/ids.go` | 领域 ID 生成 | Services |
| `pkg/collections/strings.go` | 标签集合处理 | GameService/ReportService |
| `pkg/collections/numbers.go` | 排序和数值聚合工具 | ReportService |
| `pkg/text/normalize.go` | 现场文字清理 | GameService/ReportService |
| `pkg/text/labels.go` | 玩家策略风格解释 | ReportService |
| `pkg/metrics/summary.go` | 均值、中位数、波动、趋势和相关性计算 | 分析、资源轨迹、比较和评分卡 |
| `internal/clock/clock.go` | 时间边界和日期工具 | 服务与任务 |
| `internal/telemetry/log.go` | 结构化运行日志 | HTTP/Worker/Main |
| `internal/telemetry/metrics.go` | 请求和失败计数 | HTTP Middleware/Health |
| `internal/security/password.go` | 密码散列校验 | AuthService |
| `internal/security/token.go` | HMAC 会话票据 | AuthService/Middleware |
| `internal/security/cookie.go` | 浏览器会话 Cookie | Auth API |
| `internal/service/auth.go` | 注册、登录、会话恢复 | HTTP Auth -> Store/Security |
| `internal/service/game.go` | 桌游档案和规则变体 | Games API -> Store/Domain |
| `internal/service/match.go` | 创建对局、席位、状态变更 | Matches API -> Store/Domain/Jobs |
| `internal/service/turn.go` | 开启、关闭回合、轮换玩家 | Turns API -> Match/Store |
| `internal/service/event.go` | 记录行动和高影响事件 | Events API -> Domain/Resources/Store |
| `internal/service/resources.go` | 资源账本和下限规则 | EventService -> Game/Store |
| `internal/service/milestone.go` | 关键时刻查询和标记 | Timeline/Report |
| `internal/service/timeline.go` | 回合时间线聚合 | Timeline API |
| `internal/service/replay.go` | 从事实事件重放状态 | Replay API |
| `internal/service/report.go` | 玩家线、赢家、策略摘要 | Report API/Worker |
| `internal/service/export.go` | Markdown/JSON 个人导出 | Export API |
| `internal/service/search.go` | 历史对局和桌游搜索 | Search API |
| `internal/service/analysis.go` | 从回合和行动事实生成玩家画像、资源轨迹、阶段分析和策略信号 | `/api/v1/matches/{id}/analysis` -> Store/Domain/Metrics |
| `internal/service/compare.go` | 多局对局比较、共同标签和指标差异 | `/api/v1/compare` -> MatchService/Store |
| `internal/service/reflection.go` | 复盘问题、回答、主题提取和完成率 | `/api/v1/matches/{id}/reflections` -> Store |
| `internal/service/planning.go` | 根据对局状态和数据质量生成下一步行动建议 | `/api/v1/matches/{id}/suggestions` -> Analysis |
| `internal/service/scorecard.go` | 记录完整度、节奏、资源、关键时刻和复盘可用性评分 | `/api/v1/matches/{id}/scorecard` -> Analysis |
| `internal/service/catalog.go` | 为前端行动录入提供事件类型、标签、资源和规则目录 | `/api/v1/matches/{id}/catalog` -> Store/Game/Match |
| `internal/service/insights.go` | 汇总战报、分析、评分卡和资源轨迹为可读洞察页面数据 | `/api/v1/matches/{id}/insights` -> Report/Analysis |
| `internal/service/result.go` | 泛型错误结果和空列表构造，避免服务层重复创建零值 | 各领域服务错误分支与列表初始化 |
| `internal/service/jobs.go` | 战报任务排队与重试状态 | MatchService -> Worker |
| `internal/jobs/worker.go` | 消费任务并构建战报 | Scheduler -> ReportService |
| `internal/jobs/scheduler.go` | 周期性执行后台任务 | Main -> Worker |
| `internal/jobs/retry.go` | 任务退避窗口和清理 | 运维调用/Worker |
| `internal/httpapi/app.go` | 组装 HTTP 应用 | Main -> Handler |
| `internal/httpapi/routes.go` | 注册公开和受保护路由 | App |
| `internal/httpapi/middleware.go` | 会话鉴权和用户上下文 | Protected API |
| `internal/httpapi/response.go` | JSON 响应和错误状态 | 全部 Handler |
| `internal/httpapi/decode.go` | 请求 JSON 解码 | 全部写接口 |
| `internal/httpapi/auth.go` | 认证 API | Browser -> AuthService |
| `internal/httpapi/games.go` | 桌游 API | Browser -> GameService |
| `internal/httpapi/matches.go` | 对局、回合、行动 API | Browser -> MatchService |
| `internal/httpapi/timeline.go` | 时间线和重放 API | Browser -> MatchService |
| `internal/httpapi/reports.go` | 战报 API | Browser -> ReportService |
| `internal/httpapi/search.go` | 搜索 API | Browser -> SearchService |
| `internal/httpapi/analysis.go` | 分析、比较、评分卡、行动目录、洞察和复盘 API | Browser -> Match/Reflection Services |
| `internal/httpapi/export.go` | 下载 API | Browser -> ExportService |
| `internal/httpapi/health.go` | 健康与指标 API | Browser/Monitor |
| `internal/httpapi/frontend.go` | 嵌入前端静态资源 | Browser -> web.Files |
| `web/assets.go` | 嵌入 HTML/CSS/JS | FrontendHandler |

原始需求附件未复制到项目目录。
