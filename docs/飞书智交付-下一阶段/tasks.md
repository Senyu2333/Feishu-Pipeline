# 飞书智交付 · 下一阶段任务拆解（tasks）

## 1. 使用方式

本文档把 `spec.md` 中的下一阶段范围拆解为可执行任务。后续开发时，Agent 应按本文档顺序推进，并实时更新勾选状态。

状态约定：

- `[ ]` 未开始
- `[~]` 进行中
- `[x]` 已完成
- `[!]` 阻塞或需要重新确认

本次文档更新已完成的基线核验：

- [x] 已阅读赛题文档、产品技术设计、第一阶段计划、第二阶段计划
- [x] 已核对后端 Pipeline 模型、Service、Repository、Engine、Stage Handler、Controller、Router
- [x] 已核对前端工作流、审批、交付页面当前仍以静态 Demo 数据为主
- [x] 已执行 `cd apps/api-go && go test ./...`，当前后端测试通过
- [x] 已确认下一阶段从 S3：Agent 可插拔、调用可观测、交付可查询 开始推进

---

## 2. 阶段总览

当前下一阶段建议拆为 6 个模块：

1. S3-1：Agent Provider Adapter
2. S3-2：Prompt Registry 与阶段 Agent 化
3. S3-3：AgentRun 可观测增强
4. S3-4：GitDelivery 基础交付闭环
5. S3-5：工作台聚合 API 增强
6. S3-6：Swagger、测试与演示验收

推荐顺序：

```text
S3-1 -> S3-2 -> S3-3 -> S3-4 -> S3-5 -> S3-6
```

原因：

- Provider Adapter 是阶段 Agent 化的前提
- AgentRun 可观测依赖 Agent 调用链路
- GitDelivery 依赖 delivery 阶段输出稳定
- 工作台增强依赖前面数据结构成型
- Swagger 和测试放在最后做整体收口

---

## 3. S3-1：Agent Provider Adapter

### 3.1 目标

建立统一 Agent Provider 抽象，使 Pipeline 阶段可以在 deterministic provider 和真实 LLM provider 之间切换。

### 3.2 范围

- Provider 接口定义
- Provider 配置结构
- deterministic provider 适配
- 至少一个真实 provider 的接入点或最小实现
- provider 选择逻辑
- fallback 策略

### 3.3 子任务

- [x] S3-1.1 梳理当前 Agent/AI 相关代码
  - [x] 检查 `apps/api-go/internal/agent`
  - [x] 检查 `apps/api-go/internal/external/ai`
  - [x] 确认是否已有 provider/client 可复用
  - [x] 避免重复造轮子
  - 结论：`internal/agent` 当前偏 Session 需求拆解 workflow；`internal/external/ai/ark.go` 可作为真实 provider 的包装入口；Pipeline Engine 当前直接写入 `internal/deterministic` AgentRun，下一步应抽出统一 Agent Runner。

- [x] S3-1.2 定义 Provider 接口
  - [x] 新增或复用 agent provider package
  - [x] 定义统一请求结构
  - [x] 定义统一响应结构
  - [x] 字段至少包含 prompt、input JSON、model、temperature、max tokens、metadata
  - [x] 响应至少包含 content、raw JSON、token usage、latency、finish reason
  - 结论：新增 `internal/pipeline/provider.go`，以 `AgentProvider` / `AgentProviderRequest` / `AgentProviderResponse` 统一真实模型与 fallback 观测。

- [x] S3-1.3 实现 deterministic provider
  - [x] 把当前 Stage Handler 的确定性逻辑纳入 fallback 路径
  - [x] 确保无外部配置时默认可运行
  - [x] AgentRun 中明确记录 provider=`internal` 或 `deterministic`
  - 结论：无 provider、provider 错误、JSON 解析失败、schema 校验失败时均回退 deterministic handler，并在 AgentRun tokenUsageJSON 中记录 fallbackReason。

- [~] S3-1.4 设计真实 provider 配置
  - [x] 支持 provider name
  - [x] 支持 model name
  - [x] 支持 API Key 从环境变量读取
  - [x] 支持超时配置
  - [x] 支持 demo mode 开关
  - 结论：现有 `ai.provider`、`ai.ark.*` 配置已被 Pipeline provider adapter 复用；demo mode 目前表现为“无 key 自动 fallback”，后续可补显式 `demo_mode` 配置项。

- [x] S3-1.5 接入至少一个真实 provider 的最小调用链路
  - [x] 如果使用已有 AI client，则包装为 Provider
  - [x] 如果暂不直接调用外部 API，则至少完成接口和配置占位
  - [x] 外部调用失败时返回结构化错误
  - [x] 不把 API Key 写入数据库或日志
  - 结论：启动时如 Ark client 可用，会通过 `TextGenerationProvider` 接入 Pipeline AgentRunner；API Key 仍只从配置/环境变量进入 client，不进入 AgentRun。

- [x] S3-1.6 Provider fallback 策略
  - [x] 无 API Key 时 fallback deterministic
  - [x] provider 调用失败时记录失败原因
  - [x] 明确哪些错误允许 fallback，哪些错误应中断阶段
  - [x] fallback 行为写入 AgentRun metadata 或 output

### 3.4 验收标准

- [x] 可以通过配置选择 deterministic provider
- [x] 可以通过配置选择真实 provider 或真实 provider 占位
- [x] 无 key 时 Pipeline 仍可跑通
- [x] provider 错误不会导致数据丢失
- [x] 测试覆盖 provider 选择和 fallback

---

## 4. S3-2：Prompt Registry 与阶段 Agent 化

### 4.1 目标

将各 Stage Handler 的执行逻辑升级为“Prompt + Provider + 输出校验 + fallback”的结构。

### 4.2 范围

覆盖以下阶段：

- requirement_analysis
- solution_design
- code_generation
- test_generation
- code_review
- delivery

Checkpoint 阶段仍由引擎和人工审批控制，不走 provider。

### 4.3 子任务

- [x] S3-2.1 定义 AgentKey
  - [x] requirement analyst agent
  - [x] solution designer agent
  - [x] code generator agent
  - [x] test generator agent
  - [x] code reviewer agent
  - [x] delivery integrator agent

- [x] S3-2.2 建立 Prompt Registry
  - [x] 每个阶段一个 system prompt
  - [x] 每个阶段一个 user prompt builder
  - [x] prompt 中明确输入 JSON
  - [x] prompt 中明确输出 JSON 字段
  - [x] prompt 中要求不得输出非 JSON 或需包裹可解析 JSON

- [~] S3-2.3 建立输出结构校验
  - [x] 校验 required fields
  - [ ] 校验字段类型
  - [x] 校验空输出
  - [x] 校验 JSON 解析失败
  - [x] 输出不合法时记录错误并 fallback 或 fail
  - 说明：本轮先做 required/empty/JSON parse 校验；字段类型细化留到下一批。

- [x] S3-2.4 改造 RequirementAnalysisHandler
  - [x] 构造需求分析输入
  - [x] 调用 Agent Runner
  - [x] 解析结构化需求输出
  - [x] fallback 到当前 deterministic 输出

- [x] S3-2.5 改造 SolutionDesignHandler
  - [x] 注入代码库上下文
  - [x] 调用 Agent Runner
  - [x] 输出影响文件、API 改动、数据模型改动、实现计划
  - [x] fallback 到当前 deterministic 输出

- [x] S3-2.6 改造 CodeGenerationHandler
  - [x] 输入技术方案和上下文
  - [x] 输出 changedFiles、patches、diffSummary
  - [x] 保持 patch 为受控结构，不直接危险写文件
  - [x] fallback 到当前 deterministic 输出

- [x] S3-2.7 改造 TestGenerationHandler
  - [x] 输入变更集和需求
  - [x] 保留受控命令执行
  - [x] Agent 可生成 testPlan，命令执行仍由后端白名单控制
  - [x] fallback 到当前 deterministic 输出
  - 说明：`test_generation` provider 输出不会覆盖后端白名单执行出的 commands、commandResults、status。

- [x] S3-2.8 改造 CodeReviewHandler
  - [x] 输入方案、变更集、测试结果
  - [x] 输出 conclusion、issues、securityNotes、maintainabilityNotes
  - [x] fallback 到当前 deterministic 输出

- [x] S3-2.9 改造 DeliveryHandler
  - [x] 输入评审结果和变更集
  - [x] 输出 prTitle、prBody、manualReleaseNotes、validation
  - [x] 为 GitDelivery 创建准备数据

### 4.4 验收标准

- [x] 每个可执行阶段都有 AgentKey
- [x] 每个可执行阶段都有 prompt 定义
- [x] 每个可执行阶段输出仍符合现有 schema
- [x] provider 关闭时测试仍通过
- [x] provider 开启时可记录真实调用结果或结构化错误

---

## 5. S3-3：AgentRun 可观测增强

### 5.1 目标

让每一次阶段 Agent 执行都可追踪、可审计、可展示。

### 5.2 子任务

- [x] S3-3.1 检查 AgentRun 模型字段
  - [x] provider
  - [x] model
  - [x] promptSnapshot
  - [x] inputJSON
  - [x] outputJSON
  - [x] tokenUsageJSON
  - [x] latencyMS
  - [x] status
  - [x] errorMessage
  - 结论：字段已存在；本轮已接入写入 promptSnapshot、tokenUsageJSON、latencyMS、fallback reason。

- [~] S3-3.2 统一 AgentRun 创建时机
  - [ ] 调用 provider 前记录 input/prompt
  - [x] 调用成功后记录 output/token/latency/status
  - [x] 调用失败后记录 error/status/latency
  - 说明：当前为阶段执行结束后由 Engine 一次性落库 AgentRun，避免状态写入分散；如后续需要实时工作台展示运行中 AgentRun，再补 pre-create/update。

- [~] S3-3.3 增强 token usage 记录
  - [~] input tokens
  - [~] output tokens
  - [~] total tokens
  - [~] provider 原始 usage 字段
  - 说明：结构已预留并写入 tokenUsageJSON；现有 Ark client wrapper 暂未暴露真实 usage，待 provider client 增强。

- [x] S3-3.4 增强错误记录
  - [x] provider error
  - [x] JSON parse error
  - [x] schema validation error
  - [x] fallback reason

- [x] S3-3.5 增加 service 层聚合支持
  - [x] timeline 返回 AgentRun 已可用，确认字段完整
  - [x] current 返回当前阶段最新 AgentRun
  - [x] 确认现有 AgentRun 查询 API 是否足够
  - 说明：已有 `GET /api/pipeline-runs/:id/agent-runs`，本轮未新增单条 AgentRun 详情 API。

### 5.3 验收标准

- [x] 每个执行阶段至少有一条 AgentRun
- [x] 失败阶段也有 AgentRun 错误记录
- [x] timeline/current 能展示当前阶段 AgentRun
- [x] 测试覆盖成功、失败、fallback 三类情况

---

## 6. S3-4：GitDelivery 基础交付闭环

### 6.1 目标

让 delivery 阶段从“只生成交付摘要”升级为“创建可查询的 GitDelivery 交付记录”。

### 6.2 子任务

- [x] S3-4.1 检查 GitDelivery 模型与 repository 方法
  - [x] 确认字段是否覆盖 runID、branch、commit、PR/MR URL、status、summary
  - [x] 补充 create/list/get/update 方法
  - 结论：`GitDelivery` 已补 PR/MR 标题、正文、changedFilesJSON、validationJSON；repository 已补 create/list/get/updateStatus。

- [x] S3-4.2 DeliveryHandler 生成交付草稿
  - [x] prTitle
  - [x] prBody
  - [x] changedFiles
  - [x] validation summary
  - [x] manual release notes

- [x] S3-4.3 Pipeline Engine 在 delivery 阶段创建 GitDelivery
  - [x] 与 Artifact 创建保持一致性
  - [x] 不执行 push
  - [x] 不调用远程 Git API
  - [x] status 使用 draft/ready 等安全状态

- [x] S3-4.4 新增 GitDelivery 查询 API
  - [x] `GET /api/pipeline-runs/:id/deliveries`
  - [x] `GET /api/git-deliveries/:deliveryID`
  - [x] 返回交付摘要、PR 草稿、状态、关联 run

- [x] S3-4.5 timeline 关联最新 GitDelivery
  - [x] summary 增加 latestDeliveryID
  - [x] timeline 返回 deliveries 或 latestDelivery
  - [x] current 在 delivery 阶段展示交付信息

### 6.3 验收标准

- [x] delivery 阶段完成后有 GitDelivery 记录
- [x] 可通过 API 查询交付记录
- [x] timeline 能显示交付结果
- [x] 不执行远程 push/PR 创建
- [x] 测试覆盖 delivery 创建与查询

---

## 7. S3-5：工作台聚合 API 增强

### 7.1 目标

让前端工作台可以用少量 API 渲染完整 Pipeline 状态和下一步动作。

### 7.2 子任务

- [x] S3-5.1 增强 timeline summary
  - [x] totalStages
  - [x] completedStages
  - [x] failedStages
  - [x] waitingApproval
  - [x] currentStageKey
  - [x] latestArtifactID
  - [x] latestDeliveryID
  - [x] startedAt / finishedAt / durationMs

- [x] S3-5.2 增强 current 响应
  - [x] 当前 run
  - [x] 当前 stage
  - [x] 当前 artifact
  - [x] 当前 checkpoint
  - [x] 当前 agentRun
  - [x] 当前 delivery
  - [x] nextAction

- [x] S3-5.3 设计 nextAction
  - [x] `start_run`
  - [x] `wait_execution`
  - [x] `approve_checkpoint`
  - [x] `inspect_failure`
  - [x] `review_delivery`
  - [x] `completed`
  - [x] `resume_run`
  - [x] `terminated`

- [x] S3-5.4 测试聚合逻辑
  - [x] draft run
  - [x] running run
  - [x] waiting approval run
  - [x] failed run
  - [x] completed delivery run

### 7.3 验收标准

- [x] 前端可以通过 timeline/current 判断页面主状态
- [x] waiting approval 时能拿到 checkpoint 和审批动作提示
- [x] delivery 完成后能拿到交付记录
- [x] 测试覆盖关键状态

---

## 8. S3-6：Swagger、测试与演示验收

### 8.1 目标

完成文档、测试和演示链路收口。

### 8.2 子任务

- [x] S3-6.1 Swagger 注释更新
  - [x] provider 相关接口，如有，本轮无新增 provider HTTP API
  - [x] GitDelivery 查询接口
  - [x] timeline/current 新字段
  - [x] checkpoint 审批接口保持准确

- [x] S3-6.2 生成或校验 API 文档
  - [x] 确认 docs/api 或 swagger 输出可用
  - [x] 确认新增类型在 OpenAPI 中出现

- [x] S3-6.3 后端测试
  - [x] `cd apps/api-go && go test ./...`
  - [x] provider fallback 测试
  - [x] stage agent 化测试
  - [x] delivery API 测试
  - [x] timeline/current 聚合测试

- [ ] S3-6.4 演示脚本核验
  - [ ] 创建 PipelineRun
  - [ ] start
  - [ ] 到达方案审批
  - [ ] approve
  - [ ] 到达评审确认
  - [ ] approve
  - [ ] delivery 完成
  - [ ] 查询 timeline/current/delivery

- [x] S3-6.5 更新 checklist
  - [x] 把已完成项勾选
  - [x] 标记仍未完成项
  - [x] 记录测试结果

### 8.3 验收标准

- [x] 全量测试通过
- [x] API 文档没有明显缺失
- [x] demo mode 无外部 key 可跑通
- [x] 如果配置真实 provider，错误可观测且不破坏数据
- [x] 文档与实现状态一致

---

## 9. 推荐第一批开发切片

为了降低风险，建议下一次实际编码先做第一批：

```text
S3-1 Agent Provider Adapter
S3-2 Prompt Registry 基础骨架
S3-3 AgentRun 可观测增强的最小闭环
```

第一批不急于实现 GitDelivery API。原因：

- provider 和 AgentRun 是后续所有真实 AI 能力的基础
- 当前 delivery 仍可使用 deterministic 输出支撑演示
- 先把 Agent 调用链路稳定下来，再做交付记录更安全

第一批完成后，再做第二批：

```text
S3-4 GitDelivery 基础交付闭环
S3-5 工作台聚合 API 增强
S3-6 Swagger、测试与演示验收
```
