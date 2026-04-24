# new-api v0.2 产品需求文档（PRD）

> **版本**：v0.2（本文件）
> **日期**：2026-04-24
> **状态**：Draft — 分 5 段交付；T1 ✅ / T2-T5 🟡
> **代码主干**：`Calcium-Ion/new-api` mainline（**不**在 `nexus/` 子模块上继续开发）
> **工作分支**：`claude/draft-product-requirements-hUHXs`
> **前置文件**：`CLAUDE.md`、`AGENTS.md`
> **v0.1 归档**：git `c95fc09`（"nexus-api 骨架"版本，Agent/MCP/RAG/Tenant/x402 已冻结，见 §2.4）

---

## 版本演进说明

| 版本 | 方向 | 结论 |
|---|---|---|
| v0.1 | 在 `nexus/` 新建并列 Go module，从 0 重写，含 Agent/MCP/RAG/多租户/x402 共 7 大新机制 | 发现：范围过大 + 需重写 40+ provider 适配器 → **推倒** |
| **v0.2**（本文件）| **就地在 `new-api` 主干做减法 + 增强**；C 端优先；砍 Agent/MCP/RAG；`nexus/` 资产冻结 | **采纳** |

---

## 文档结构导引

| 段 | 章节 | 状态 | 读者 |
|---|---|---|---|
| **T1** | §1-§2 产品定位 + 范围冻结 | ✅ 本段 | 所有人（先读） |
| T2 | §3 功能需求（8 模块） | 🟡 待写 | PM / 工程 |
| T3 | §4-§6 技术架构 + 数据模型 + API 契约 | 🟡 待写 | 工程 |
| T4 | §7-§9 非功能需求 + 4 周路线图 + 接受标准 | 🟡 待写 | PM / QA |
| T5 | §10-§11 + 附录 | 🟡 待写 | Claude Code 执行用 |

**阅读顺序建议**：产品经理只需 T1+T2+T4；Claude Code 编码时读 T1+T2+T3+T5。

---

## §1 产品定位

### 1.1 产品名

对外名称沿用 **new-api**（不改 repo 名以保留社区资产），v0.2 版本线对外称 **"new-api Pro Gateway"**。最终中文副标题在 T5 决策日志里拍板。

### 1.2 一句话定义

> **为个人开发者和小团队提供"可观测 + 多渠道采购 + 加密货币结算"的 AI API 统一网关。**

### 1.3 用户画像

**主画像 A — 自由职业 / 独立开发者（C 端核心）**
- 同时用 3–8 个 API key（OpenAI 官方 / 第三方中转 / Claude / Gemini / 国产模型 …）
- 月支出 $20–500 USD
- 痛点：记不清哪 key 剩多少额 / 哪家最近频繁超时 / 哪家 prompt cache 命中率最高
- 为什么选我们：统一 token 观测仪表盘 + 自动熔断换路 + 稳定币一次充值打通多家上游

**主画像 B — 3–10 人技术小团队 / 初创公司 CTO（C 端延伸）**
- 想给全员统一发 API token，避免每个人各自管 key
- 需要按人 / 项目看成本
- 为什么选我们：团队共享渠道池 + 成员级观测 + 成本审计

**次画像 C — 企业技术管理部门（B 端，v0.3+）**
- 公司采购了多模型厂商的 key，要按部门分发
- 需要多租户 / RBAC / 审计
- v0.2 **不交付**，但数据层（`OrgID` 字段）**预留**，避免 v0.3 推倒表结构

### 1.4 竞品对标

| 竞品 | 它的强项 | v0.2 差异化 |
|---|---|---|
| **OpenRouter** | 多渠道聚合、自动路由 | + OSS 自部署 / 加密货币结算 / 本地化观测 |
| **Portkey Gateway** | 熔断 / 重试 / 负载均衡策略丰富 | 吸收其熔断策略；+ 观测仪表盘 + Web3 登录（见 `github.com/portkey-ai/gateway`） |
| **Helicone** | 请求级观测 / 代理层易接入 | + 多渠道采购，不只是 observability |
| **Langfuse** | trace / eval 最全的 LLM observability | 我们不做 eval，只做 token 级 metrics + 请求详情（见 `github.com/langfuse/langfuse`） |
| **new-api 原版** | 中文社区、40+ provider、充值体系成熟 | 砍 playground，加 Web3/稳定币/熔断/prompt-cache 感知 LB |

**核心差异公式**：**"v0.2 = Portkey 的路由策略 + Langfuse 的观测视图 + new-api 的 provider 覆盖 + Web3/稳定币结算"**。

### 1.5 商业模型（混合定位 C，v0.1 §三·硬伤 3 已确认）

- **OSS 本体免费**：GitHub 公开，用户自部署 0 费用。
- **官方托管 SaaS**：付 SaaS 订阅费 **或** 按 token 消耗付费。
- **Key 供应策略**：
  - **BYO-Key（用户自带 key）**：平台只收 SaaS 费，价值在于观测 + 熔断 + LB。
  - **Platform-Key（平台自营 key）**：平台按量加价转售。
  - 在 Token / Channel 管理 UI 上**显式区分**两类 key，用户可按请求级选择走哪类。
- **Token 交易所（用户 P2P 出售闲置 key）**：v0.3+ 研究，v0.2 **不做**。
- **法律定位文案**：官网 landing 定位 **"API 管理与观测平台"**（合规表述），ToS 写明"用户自行承担上游 ToS 合规责任"。具体措辞 T5 拍板。

### 1.6 北极星指标（North Star Metric）

**v0.2 GA 后 60 天内目标**：

| 指标 | 目标值 | 测量方式 |
|---|---|---|
| 自部署用户数（GitHub fork + docker pull） | 500+ | GitHub Insights + Docker Hub |
| SaaS 注册用户数 | 200+ | 官方版 `users` 表 |
| SaaS 日活（至少发 1 次 relay） | 50+ | `logs` 表 distinct user_id |
| 稳定币充值完成笔数 | 30+（POC 阶段） | 新 `stablecoin_payments` 表 |
| Web3 登录成功率 | ≥ 95% | 登录埋点 |
| 熔断器误触率 | ≤ 1%（正常 channel 被错误熔断的比例） | 熔断事件日志 |

**北极星指标** = **SaaS 日活用户数（DAU）**。功能优先级按"是否提升 DAU 留存"排序。

---

## §2 v0.2 范围冻结（Scope Lock）

> **冻结原则**：任何不在本章节明确列出的功能，v0.2 一律不做。
> 变更需走 §2.6 范围变更流程。

### 2.1 继承（Inherited — 从 new-api 主干**不改动**移植）

| 模块 | 源文件锚点 | 说明 |
|---|---|---|
| 40+ Provider 适配器 | `relay/channel/*` | 原样保留 |
| OpenAI / Claude / Gemini / Responses 协议兼容 | `relay/`、`controller/relay.go`、`service/openai_chat_responses_*.go` | 原样 |
| 邮箱 + 密码注册 | `controller/user.go`、`model/user.go` | 原样 |
| OAuth（GitHub / Discord / 自定义 OIDC） | `controller/oauth.go`、`controller/custom_oauth.go`、`model/custom_oauth_provider.go` | 原样 |
| Passkey / WebAuthn | `controller/passkey.go`、`model/passkey.go`、`service/passkey/` | 原样 |
| 邀请码 / 兑换码 / 礼品码 | `controller/redemption.go`、`model/redemption.go` | 原样 |
| 签到 / 任务 | `controller/checkin.go`、`model/checkin.go`、`model/task.go` | 原样 |
| 法币充值（epay / Stripe / 支付宝 / 微信） | `controller/billing.go`、`service/epay.go`、`controller/topup.go`、`router/api-router.go:48,63,83-89,114,137-161` | 原样 |
| Midjourney 异步任务 | `controller/midjourney.go`、`model/midjourney.go`、`relay/channel/midjourney/` | 原样 |
| Suno 异步任务 | `relay/channel/suno/` + 相关 controller | 原样 |
| 6 国语言 i18n（zh/en/fr/ru/ja/vi） | `i18n/`、`web/src/i18n/locales/*.json` | 原样；v0.2 新增文案必须补齐 6 国翻译 |
| 渠道亲和缓存（现有 783 行） | `service/channel_affinity.go`、`controller/channel_affinity_cache.go` | **原样保留**；N2 在此基础上叠加 prompt-cache 感知，不重写 |
| 管理后台其它页面（Dashboard / Log / Pricing / Token / Channel / Model / Redemption / TopUp / Subscription 等） | `web/src/pages/*` + 对应 `controller/*` | 原样 |

### 2.2 新增（Net-New — v0.2 必须交付）

| # | 功能 | 主要落地位置 | 优先级 |
|---|---|---|---|
| **N1** | **熔断器**（Circuit Breaker，滑动窗口错误率 + 连续超时双触发 + 指数退避冷却 + 半开探测）| 新 `service/circuit_breaker.go` + 扩展 `middleware/distributor.go` | P0 |
| **N2** | **Prompt-cache 感知 LB**（相同前缀 → 同一 channel 一致性哈希，最大化 Claude/OpenAI prompt caching 命中） | 扩展 `service/channel_select.go`、`service/channel_affinity.go` | P0 |
| **N3** | **Token 级可观测仪表盘**（QPS / 延迟 / 成本 / 错误率 / token 消耗分布 + 请求详情懒加载） | 新 `controller/observability.go` + 新 `web/src/pages/Observability/` | P0 |
| **N4** | **EVM 钱包登录**（Metamask / WalletConnect / Coinbase Wallet / Rabby 等所有 EVM 钱包；SIWE EIP-4361） | 新 `controller/wallet_auth.go` + 新 `model/wallet_credential.go` + 新 `web/src/pages/Login/WalletConnect.tsx` | P0 |
| **N5** | **稳定币充值**（USDT / USDC / JPYP 预充值，多链） | 新 `service/stablecoin/` + 新 `controller/stablecoin.go` + 新 `model/stablecoin_payment.go` | P0 |
| **N6** | **平台自营 Key 池 UI**（Channel 页区分 BYO vs Platform，Token 页允许用户勾选"允许走平台 key"） | 扩展 `model/channel.go`、`model/token.go` + 前端 Channel/Token 页改造 | P1 |
| **N7** | **`OrgID` 字段预留**（`users / tokens / channels / logs` 四张表新增可空 `org_id`） | `model/*.go` + `AutoMigrate` | P1（B 端铺垫，不暴露 UI） |
| **N8** | **官网 Landing + 部署文档 + OSS 发布流程** | 新 `web/src/pages/Landing/` + `docs/deployment/` + `README.md` 更新 + GitHub Release CI | P0 |

**优先级定义**：
- **P0** = v0.2 GA 必须；不完成则推迟发布。
- **P1** = v0.2 GA 希望；若排期吃紧可降级为"功能入口留白 + v0.2.1 补"。

### 2.3 删除（Removed — v0.2 必须从代码库移除 / 前端隐藏）

| # | 功能 | 移除方式 |
|---|---|---|
| **D1** | `/pg/*` 后台 Playground 对话路由 | 删除 `router/relay-router.go:59-65` + `controller/playground.go` |
| **D2** | `web/src/pages/Playground/*` 前端页面 | 删除目录 + `App.jsx` 路由 + 菜单项 |
| **D3** | `web/src/pages/Chat/*` 前端页面 | 删除目录 + 路由 + 菜单项 |
| **D4** | `web/src/pages/Chat2Link/*` 前端页面 | 删除目录 + 路由 |
| **D5** | 侧边栏"聊天"菜单项 | `web/src/components/layout/SiderBar.jsx` 清理 |

> **注意**：只删前端对话页 + `/pg` 后端路由。**不动** Midjourney / Suno（它们是"异步任务"类，不是"聊天对话"类）。

### 2.4 冻结 / 延后（Parked — v0.2 **不投产**，但保留代码占位）

| # | 资产 | 状态 | 归宿 |
|---|---|---|---|
| **P1** | `nexus/` 子模块全部代码（58 个文件，1576 行） | **原地冻结**，不合并进主干；补 "v0.2 parked" 说明 | v0.3+ 评估是否复用 |
| **P2** | Solana / Phantom 钱包登录 | 代码冻结在 `nexus/internal/auth/web3/solana.go` | v0.3+ |
| **P3** | Coinbase OAuth | **不做** | — |
| **P4** | x402 协议按次付费 | 骨架冻结 | v0.2.x 稳定币预充值跑通后评估 |
| **P5** | 多租户 / 组织 UI / RBAC 中间件 | `OrgID` 字段保留（N7），**UI + Controller 不做** | v0.3 B 端 |
| **P6** | Token 交易所（用户 P2P 卖 key） | 0 代码 | v0.3+ 研究 |
| **P7** | Agent / MCP / RAG / KnowledgeBase | **v0.2 完全不做** | 独立产品线，非本项目范围 |

### 2.5 明确的非目标（Non-Goals）

以下 v0.2 **明确不支持**，防止需求蔓延：

- ❌ **企业 SSO / SAML / LDAP**（等 B 端）
- ❌ **私有化部署服务合同 / 按年 license**（等 B 端）
- ❌ **模型训练 / 微调托管**
- ❌ **Eval / 自动评测**（Langfuse 做的事不做）
- ❌ **Prompt 管理 / Prompt 版本控制**（Portkey 做的事不抄）
- ❌ **自研前端编辑器 / Workflow 画布**
- ❌ **Mobile App（iOS / Android 原生）**
- ❌ **Agent 编排 / 工具调用 / MCP / RAG**（见 §2.4 P7）

### 2.6 范围变更流程

v0.2 开工后任何范围调整必须走：

1. **你发起**：对话里写 "范围变更：XX 从 Parked 挪到新增 / 从 Removed 移回继承"。
2. **我评估**：是否影响 4 周排期？是否影响其它模块？是否影响已 commit 的代码？
3. **双方确认**后，我更新本 PRD §2 并单独 commit（commit 名 `docs(prd): scope change — ...`）。
4. **未走流程的口头改动不纳入**。Claude Code 后续编码只信任 PRD 里写死的东西。

---

_T1 结束。等你 ✅ 后开始 T2（§3 功能需求，8 个模块逐一细化到用户故事 + 验收标准）。_
