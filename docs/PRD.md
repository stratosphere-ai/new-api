# nexus-api 产品需求文档（PRD）

> **版本**：v0.1 — 上半部分（现状梳理）
> **日期**：2026-04-22
> **状态**：Draft（下半部分待补充：目标功能细节）
> **代码协作对象**：Claude Code（本文件用于向 Claude Code 提供上下文）
> **对应分支**：`claude/draft-product-requirements-hUHXs`

---

## 0. 文档说明

本文档是 `new-api` 仓库中 **下一代后端骨架 `nexus-api`** 的产品需求文档。

- **上半部分（本次交付）**：梳理项目背景、现有 `new-api` 的能力盘点、`nexus/` 目录下已经完成的 Phase 1 骨架工作、目录与接口定义清单、技术栈与工程约束。
- **下半部分（待补充）**：具体要实现的功能需求、优先级、验收标准、分阶段路线图。

> 本文件用于向 Claude Code 注入项目上下文。在后续编码任务中，Claude Code 应**严格遵循** `CLAUDE.md` / `AGENTS.md` 里的工程规则（JSON 封装、三库兼容、Bun 前端等），并以本文件列出的骨架文件作为"锚点"进行增量实现。

---

## 1. 项目背景

### 1.1 `new-api`：已经存在的产物

`new-api` 是一个基于 Go 的 **AI API 网关 / 聚合代理**，在一个统一 API 后面聚合 40+ 上游 AI 服务提供商（OpenAI、Claude、Gemini、Azure、AWS Bedrock 等），并提供用户管理、计费、限流和管理后台。

| 维度 | 现状 |
|---|---|
| 后端 | Go 1.22+ / Gin / GORM v2 |
| 前端 | React 18 / Vite / Semi Design / Bun |
| 数据库 | 同时支持 SQLite、MySQL(≥5.7.8)、PostgreSQL(≥9.6) |
| 缓存 | Redis（go-redis）+ 进程内缓存 |
| 鉴权 | JWT / WebAuthn（Passkey）/ OAuth（GitHub、Discord、OIDC 等） |
| 架构层次 | Router → Controller → Service → Model |
| i18n | 后端：go-i18n（zh/en）；前端：i18next（zh/en/fr/ru/ja/vi） |

对应源码目录（参见 `CLAUDE.md §Architecture`）：

```
router/         HTTP 路由（API、relay、dashboard、web）
controller/     请求处理器
service/        业务逻辑
model/          GORM 数据模型
relay/          AI 转发层
  relay/channel/  逐个上游 Provider 的适配器（openai/ / claude/ / gemini/ / aws/ …）
middleware/     Auth / RateLimit / CORS / Logging / Distributor
setting/        配置（ratio / model / operation / system / performance）
common/         公共工具（JSON 封装、crypto、Redis、env、rate-limit …）
dto/ constant/ types/
i18n/ oauth/ pkg/
web/            React 前端
```

### 1.2 为什么引入 `nexus-api`

`new-api` 已经覆盖了"AI 代理 + 账号 + 计费"的基础能力，但以下场景尚未原生支持：

1. **多租户（Organization / Team）**：目前所有资源（Channel、Token、Log）归属于用户；缺少"组织"这一中间层，无法做组内 RBAC、组内配额、组内路由策略。
2. **Web3 钱包登录**：缺少以太坊（SIWE / EIP-4361）和 Solana（ed25519）两条主流钱包的签名登录通道。
3. **加密货币结算**：缺少 USDT/USDC 多链（Ethereum / Tron / BSC / Solana）的充值到账路径。
4. **x402 按次付费协议**：缺少 Coinbase `x402` HTTP 402 Payment Required 协议的网关实现，无法做"不预充值、逐请求结算"。
5. **动态路由策略**：现有的 `service.CacheGetRandomSatisfiedChannel` 只做简单加权，缺少 cost / latency / quality / A-B / canary / shadow 等策略槽位。
6. **合规与审计**：缺少独立的 AuditLog（区分于计费 Log）、内容安全 Pre/Post、按组织配置的合规策略。
7. **Agent / MCP / RAG**：缺少多步工具调用编排、MCP 桥接、知识库 RAG。

因此，在 `new-api` 根目录下新增一个**同级 Go 模块** `nexus/`，以"新骨架 + 渐进移植"的方式承接上述新机制。

---

## 2. 现阶段工作内容（Phase 1 已交付）

> 对应提交：`4728d50 feat(nexus): scaffold new backend skeleton with multi-tenant + web3 + x402 foundations`
> 交付物：目录结构、路由注册占位、核心接口定义、GORM 数据模型。
> 交付状态：`go build ./...` / `go vet ./...` 通过；所有业务实现以 `// TODO: implement` 占位。

### 2.1 模块与目录结构

```
nexus/
├── README.md                 新机制概览
├── go.mod                    module github.com/stratosphere-ai/nexus-api
├── cmd/
│   ├── server/main.go        服务入口（镜像 new-api/main.go，预留 OTel / 钱包 / MCP 初始化）
│   └── migrate/main.go       AutoMigrate 入口
└── internal/
    ├── config/               运行时配置（env + yaml 分层加载）
    ├── router/               顶层路由组注册
    ├── middleware/           中间件槽位（含新增 Tenant / RBAC / X402Gate / Audit / ContentSafety / Compliance / OTel）
    ├── model/                GORM 实体 + AutoMigrate
    ├── auth/web3/            Web3 钱包登录（SIWE + Solana + Nonce + JWT）
    ├── payment/              支付结算抽象
    │   ├── settler.go        PaymentSettler 接口
    │   ├── crypto/           USDT/USDC 多链充值（Settler / Wallet / Watcher）
    │   └── x402/             Coinbase x402 协议（Verifier / RequirementsBuilder / Facilitator）
    ├── tenant/               多租户（Service / OrgResolver / WithOrgScope GORM Scope）
    ├── routing/policy/       动态路由策略（RoutingPolicy 接口 + Registry）
    ├── observability/
    │   ├── otel/             OpenTelemetry 初始化
    │   ├── audit/            AuditLogger + Redactor
    │   ├── safety/           Filter（Pre/Post 内容安全）
    │   └── compliance/       Checker（按组织加载合规策略）
    ├── agent/                Agent 编排（Orchestrator + ToolProxy）
    │   └── mcp/              Model Context Protocol 桥接
    └── knowledge/            RAG（Store + Embedder）
```

### 2.2 路由骨架

顶层 `SetRouter` 注册以下路由组（`nexus/internal/router/main.go:9-21`）：

| 前缀 | 职责 | 备注 |
|---|---|---|
| `/api/v2/*` | 控制台 API | 镜像 `new-api/router/api-router.go`，挂载 `TenantResolver + RBAC` |
| `/api/v2/org/*` | 组织管理 | 新增：创建 / 列表 / 成员 / 角色 |
| `/api/v2/auth/web3/*` | Web3 登录 | 新增：`nonce / verify / link / unlink` |
| `/pay/x402/*` | x402 支付 | 新增：`quote / settle / facilitator/callback` |
| `/agent/*` | Agent 执行 | 新增：`:id/run`（SSE）/ `:id/runs/:run_id` / `:id/tool/:name` |
| `/v1`, `/v1beta`, `/mj`, `/suno`, `/pg` | AI 转发 | 兼容 `new-api` 既有客户端 |
| `/internal/healthz` | 健康检查 | 已实装 |

中继链中间件次序（外 → 内，见 `nexus/internal/router/relay_router.go:9-23`）：

```
RequestId → Recover → CORS → I18n → Logger/Stats → OTel
  → Decompress → SystemPerformanceCheck
  → TokenAuth → TenantResolver → RBAC
  → ContentSafetyPre → ModelRequestRateLimit → CompliancePolicy
  → X402Gate → Distribute → AuditPre → controller.Relay
  → AuditPost / ContentSafetyPost (defer) → BillingPost
```

### 2.3 中间件槽位（全部为 TODO）

| 文件 | 中间件 | 来源 / 新增 |
|---|---|---|
| `middleware/common.go` | RequestID / Recover / CORS / I18n / Logger / SystemPerformanceCheck / BodyStorageCleanup | 移植 `new-api/middleware/*` |
| `middleware/auth.go` | UserAuth / AdminAuth / RootAuth / TokenAuth / Web3JWT | 移植 + 新增 `Web3JWT` |
| `middleware/tenant.go` | TenantResolver | **新增**：按 Token.OrgID / `X-Org-ID` / 子域 / 用户默认 Org 四级回退 |
| `middleware/rbac.go` | RBAC(minRole) | **新增**：owner/admin/member |
| `middleware/x402_gate.go` | X402Gate | **新增**：额度不足时返回 402 + `X-PAYMENT-REQUIRED` |
| `middleware/distributor.go` | Distribute | 改造：委托给 `routing/policy.RoutingPolicy.Pick` |
| `middleware/rate_limit.go` / `model_rate_limit.go` | RateLimit / ModelRequestRateLimit | 移植 |
| `middleware/audit.go` | AuditPre / AuditPost | **新增** |
| `middleware/content_safety.go` | ContentSafetyPre / ContentSafetyPost（流式安全） | **新增** |
| `middleware/compliance.go` | CompliancePolicy | **新增** |
| `middleware/otel.go` | OTel tracing | **新增** |

### 2.4 数据模型（GORM）

由 `model.AutoMigrate`（`nexus/internal/model/migrate.go:8-18`）统一挂载：

| 实体 | 关键字段 | 用途 |
|---|---|---|
| `Organization` | `Slug / OwnerUserID / Plan / Settings` | 租户根节点 |
| `OrgMember` | 复合唯一 `(OrgID, UserID)` + `RoleID` | 成员关系 |
| `Role` | `OrgID / Permissions(JSON)` | `OrgID=0` 为系统模板角色 |
| `Web3Credential` | `UserID / Chain / Address(unique) / Verified / IsPrimary` | 钱包绑定 |
| `CryptoPayment` | `TxHash(unique) / Asset / Chain / AmountRaw / AmountUSD / Confirmations / Status` | 链上充值落账 |
| `X402Payment` | `PaymentIntentHash(unique) / RequestID / Status / SettlementTxID / LogID` | x402 单请求结算 |
| `SettlementTx` | `TxHash(unique) / Count / AmountUSD` | 多 X402Payment 聚合成一笔链上 TX |
| `RoutingPolicy` | `OrgID / Name / Kind(weighted\|cost\|latency\|quality\|ab\|canary\|shadow) / Rules(JSON)` | 路由策略定义 |
| `RoutingStat` | `(OrgID, ChannelID, Model, WindowStart)` 四元索引 | cost/latency/quality 策略的数据源 |
| `AuditLog` | `RequestID / PromptHash / PromptRedacted / ResponseHash / SafetyVerdict / LatencyMs` | 独立于计费 Log 的审计轨迹 |
| `CompliancePolicy` | `OrgID(unique) / Region / PIIRules / BannedModels / RetentionDays` | 组织级合规配置 |
| `AgentDefinition` | `Name / SystemPrompt / Tools(JSON) / KBIDs(JSON) / DefaultModel / Version` | 可复用 Agent 模板 |
| `AgentRun` | `AgentID / InputHash / Status / Steps(JSON) / CostQuota` | 单次 Agent 执行 |
| `ToolBinding` | `AgentID / Kind(http\|builtin\|mcp) / Config(JSON)` | 挂到 Agent 的工具 |
| `McpServer` | `OrgID / URL / AuthType / Credentials` | 外部 MCP 端点 |
| `KnowledgeBase` | `OrgID / Embedder / VectorStore(pgvector\|qdrant\|weaviate\|milvus)` | 知识库容器 |
| `KBDocument` | 复合唯一 `(KBHashOrg, Hash)` | 源文档 |
| `KBChunk` | `DocID / Seq / Text / Embedding(bytea)` | 可检索段落 |

对既有模型（`new-api/model/user.go` 等）的字段增量（见 `nexus/internal/model/user_ext.go` 占位说明）：

- `User` → 新增 `DefaultOrgID`、`Web3PrimaryAddress`
- `Token` → 新增 `OrgID`、`X402Enabled`、`X402MaxPerReqUSD`
- `Channel` / `Log` → 新增可空 `OrgID`（`null` 表示共享池 / 历史行）

### 2.5 核心接口定义（全部已声明、全部 TODO 实现）

| 包 | 接口 | 职责 |
|---|---|---|
| `auth/web3` | `Web3Authenticator` | `IssueNonce` / `Verify`，由 `SIWEAuthenticator`（EIP-4361 / secp256k1）与 `SolanaAuthenticator`（ed25519）实现 |
| `auth/web3` | `NonceStore` | Redis 存取 `(chain, address) → nonce`，带 TTL |
| `auth/web3` | `JWTIssuer` | Web3 会话 JWT（**独立于** API Token 的 `sk-*`） |
| `payment` | `PaymentSettler` | 统一支付结算抽象（Stripe / epay / CreEM / 链上 / x402 均实现此接口） |
| `payment/crypto` | `Settler / WalletService / Watcher` | HD 地址派生 / 链上订阅 / 充值落账 |
| `payment/x402` | `Verifier / RequirementsBuilder / Settler / FacilitatorClient` | x402 协议实现（见 `protocol.go:15-76`） |
| `tenant` | `OrgResolver / Service` | 租户解析 + 组织 CRUD |
| `tenant` | `WithOrgScope(orgID) func(*gorm.DB) *gorm.DB` | GORM Scope 助手，把租户边界收敛到一处 |
| `routing/policy` | `RoutingPolicy / Registry` | `Name() / Pick / Observe`；内建 `weighted / cost / latency / quality / ab / canary / shadow` 七种 |
| `observability/otel` | `Init(ctx, endpoint)` | OTLP HTTP 导出器初始化 |
| `observability/audit` | `Logger / Redactor / Entry` | 审计批写 + PII 脱敏 |
| `observability/safety` | `Filter` | `PreCheck / PostCheck`，可插 OpenAI Moderation / 自研 / 正则 |
| `observability/compliance` | `Checker / Policy` | 按 `OrgID` 加载并校验模型黑白名单、区域、保留天数 |
| `agent` | `Agent / ToolProxy` | 多步编排 + 工具执行 |
| `agent/mcp` | `Client / Tool` | MCP 协议桥接 |
| `knowledge` | `Store / Embedder / Document / Chunk` | RAG 存取与检索 |

### 2.6 工程约束（来自 `CLAUDE.md` / `AGENTS.md`）

Claude Code 在后续编码时必须遵守：

1. **JSON 封装**：业务代码 `encoding/json` 的 Marshal/Unmarshal **禁止直接调用**，一律走 `common/json.go`（`common.Marshal / Unmarshal / UnmarshalJsonStr / DecodeJson / GetJsonType`）。`json.RawMessage / json.Number` 等类型引用允许。
2. **三库兼容（SQLite / MySQL ≥ 5.7.8 / PostgreSQL ≥ 9.6）**：
   - 优先走 GORM 方法，不写原生 SQL。
   - 必须用原生 SQL 时：列引号差异用 `commonGroupCol / commonKeyCol`；布尔值用 `commonTrueVal / commonFalseVal`；按需分支 `common.UsingPostgreSQL / UsingMySQL / UsingSQLite`。
   - 禁用无回退的方言：`GROUP_CONCAT`（无 `STRING_AGG` 对应）、`@>`/`JSONB` 操作符、SQLite 不支持的 `ALTER COLUMN`、`JSONB` 直接列类型（用 `TEXT`）。
   - 新表字段用 `text` 存 JSON（见 2.4 各模型 `Permissions/Rules/PIIRules` 字段写法）。
3. **前端包管理器**：`web/` 下一律 `bun`（`bun install / run dev / run build / run i18n:*`）。
4. **i18n**：
   - 后端：`nicksnyder/go-i18n/v2`（en / zh）。
   - 前端：`i18next` + `react-i18next`，Key 用中文原文，扁平 JSON 放 `web/src/i18n/locales/*.json`，通过 `bun run i18n:extract / sync / lint` 工具链维护。

### 2.7 现状边界（明确**未做**的事）

- 所有接口为 `// TODO: implement`，服务启动后除 `/internal/healthz` 外**不提供任何实际业务响应**。
- `cmd/server/main.go` 未接入数据库、Redis、i18n、session、OTel、钱包客户端、MCP 注册表。
- `cmd/migrate/main.go` 仅打印日志，未调 `AutoMigrate`。
- `config.Load()` 仅返回 `ListenAddr: ":3000"`，未解析 env / yaml。
- `SetWebRouter` 未接入前端 embed.FS。
- `nexus/` 与 `new-api/` 根目录是**两个独立 Go module**；`nexus` 不 import `new-api`，后者的实现将按阶段**移植进 `nexus`**（而不是跨模块调用）。

### 2.8 已确认通过的校验

- `go build ./...`：`nexus/` 模块和 `new-api/` 根模块均通过。
- `go vet ./...`：`nexus/` 模块通过。
- 提交 `4728d50` 包含 58 个文件新增、合计 1576 行；diff 分布见提交正文。

---

## 3. 相关资产索引（给 Claude Code 的 "锚点"）

在执行"从 `new-api` 移植到 `nexus`"类任务时，可按下表查找对应实现：

| 需要移植的能力 | 来源（new-api） | 目标（nexus） |
|---|---|---|
| 服务入口 / 初始化 | `main.go:142-182` | `cmd/server/main.go` |
| 顶层路由 | `router/main.go` | `internal/router/main.go` |
| 转发路由 | `router/relay-router.go` | `internal/router/relay_router.go` |
| Dashboard API | `router/api-router.go` | `internal/router/api_v2_router.go` |
| 鉴权中间件 | `middleware/auth.go` | `internal/middleware/auth.go` |
| 请求 ID / CORS / Recover / Perf | `middleware/request-id.go` / `cors.go` / `performance.go` / `body_cleanup.go` | `internal/middleware/common.go` |
| 分发选路 | `service/*` + `CacheGetRandomSatisfiedChannel` | `internal/routing/policy/*` + `middleware/distributor.go` |
| 环境变量 / 配置 | `common/env.go`、`setting/system_setting/*` | `internal/config/config.go` |
| Web 前端挂载 | `router/web-router.go`（`go:embed`） | `internal/router/web_router.go` |

---

## 4. 下半部分预告（将在 v0.2 补充）

下一版将补齐以下章节，请 Claude Code 在收到 v0.2 之前**不要自行推测下半部分的需求**：

- §5 目标功能清单（按 2.5 七大方向细化到用户故事）
- §6 优先级与分阶段路线图（Phase 2/3/4）
- §7 非功能性需求（性能、可用性、安全、合规）
- §8 接口契约（REST/SSE/Webhook 格式）
- §9 验收标准与测试策略
- §10 风险与取舍

---

_本文档维护者：项目作者_
_本文档使用方：Claude Code（须与 `CLAUDE.md`、`AGENTS.md` 联合阅读）_
