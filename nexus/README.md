# nexus-api

基于 new-api 的新一代 AI API 后台骨架。此目录为 Phase 1 交付的**代码骨架**：仅包含目录结构、路由注册占位、核心接口定义。业务实现将按阶段从 `../` 根目录下的 new-api 移植并叠加新机制。

## 新机制概览

- **Web3 钱包登录**：`internal/auth/web3`（SIWE + Solana ed25519）
- **加密货币结算**：`internal/payment/crypto`（USDT/USDC 多链）
- **x402 结算协议**：`internal/payment/x402`（HTTP 402 按次付费）
- **多租户/组织**：`internal/tenant`（Organization/OrgMember/Role 三件套）
- **动态路由策略**：`internal/routing/policy`（weighted/cost/latency/quality/ab/canary/shadow）
- **可观测性/审计**：`internal/observability`（OTel/AuditLog/SafetyFilter/Compliance）
- **Agent/工作流**：`internal/agent`、`internal/knowledge`（MCP、RAG）

## 目录参考

参见 `/root/.claude/plans/new-api-new-api-validated-flurry.md` 第三部分。

## 现阶段说明

- 所有接口实现均为 `// TODO: implement` 占位。
- `go build ./...` 应可通过，但启动后仅挂载空路由。
- 业务能力按计划的 4 阶段逐步移植/实现。
