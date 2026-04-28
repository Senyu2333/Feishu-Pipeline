# Feishu Pipeline · AI 驱动的飞书研发全流程引擎

> 一款 AI Native 的研发流程自动化平台，深度集成飞书生态，打通「需求 → 方案 → 编码 → 测试 → 评审 → 交付」全链路，每个环节由专门的 AI Agent 执行，开发者只需在关键节点做决策确认。

---

## 一、项目概述

传统软件研发流程依赖人力密集型流水线：产品经理写 PRD → 技术负责人拆解任务 → 开发者编码 → 代码评审 → 测试 → 上线。每个环节依赖不同角色，信息在环节间反复丢失和变形。

**Feishu Pipeline** 将这条链路编排成 AI 驱动的 Pipeline，每个环节由专门的 AI Agent 负责执行，人类只需在关键节点做决策确认。

---

## 二、技术架构

### 2.1 技术栈

| 层级 | 技术选型 |
|------|----------|
| **核心后端** | Go 1.24 + Gin + GORM + Eino |
| **集成服务** | Node.js 20 + TypeScript + Fastify |
| **前端** | React 18 + TypeScript + Vite + Ant Design 6 + Tailwind CSS |
| **数据库** | SQLite |
| **部署** | Docker + Docker Compose |

### 2.2 项目结构

```
Feishu-Pipeline/
├── apps/
│   ├── api-go/           # 核心 Pipeline 引擎（Go）
│   │   ├── cmd/          # CLI 入口（root/serve）
│   │   ├── config/       # 配置文件
│   │   ├── internal/
│   │   │   ├── agent/    # AI Agent 定义（prompt/schema/validator/workflow）
│   │   │   ├── bootstrap/# 应用初始化（app.go/config.go）
│   │   │   ├── controller/# HTTP 控制器
│   │   │   ├── external/ # 外部集成（AI/飞书）
│   │   │   ├── job/      # 任务运行器
│   │   │   ├── model/    # 数据模型
│   │   │   ├── repo/     # 数据仓库
│   │   │   ├── router/   # 路由定义
│   │   │   ├── service/  # 业务服务层
│   │   │   └── type/     # 类型定义
│   │   └── data/         # SQLite 数据库
│   │
│   ├── api-ts/           # 飞书集成服务（TypeScript）
│   │   ├── src/
│   │   │   ├── api/
│   │   │   │   ├── document/  # 飞书文档 API
│   │   │   │   ├── table/     # 多维表格 API
│   │   │   │   └── test/      # AI 对话与 Function Calling
│   │   │   ├── lib/
│   │   │   │   ├── feishu.ts  # 飞书 SDK 封装（40+ 端点）
│   │   │   │   ├── tools.ts   # AI 工具函数（14 个工具）
│   │   │   │   └── http.ts    # HTTP 客户端
│   │   │   ├── routes/       # 路由模块
│   │   │   └── index.ts       # 服务入口
│   │   └── dist/             # 编译输出
│   │
│   └── web/              # 前端管理后台（React）
│       └── src/
│           ├── components/  # 公共组件（ChatInput/ChatMessage/RequirementCard/Sidebar/TopNav）
│           ├── lib/         # 工具库（飞书前端 SDK）
│           ├── pages/       # 页面（Home/Workflows/Approvals/Monitoring/Delivery/NewRequirement/Session/Debug）
│           └── App.tsx      # 应用入口
│
├── packages/
│   └── shared/           # 前后端共享类型与工具
│
├── deploy/               # Docker 部署配置
├── docs/                 # 项目文档
├── spec.md               # 产品需求规格说明
└── docker-compose.yml    # 容器编排
```

---

## 三、核心能力

### 3.1 Pipeline 引擎

| 能力 | 说明 |
|------|------|
| **阶段管理** | 支持 Stage 定义、排序、依赖管理 |
| **Agent 编排** | 每个阶段绑定一个或多个 AI Agent，支持 Eino 框架角色定义 |
| **数据流转** | 上一阶段输出作为下一阶段输入，支持上下文感知 |
| **生命周期** | 启动、暂停、恢复、终止全流程管理 |
| **Human-in-the-Loop** | 内置 2 个检查点（方案设计审批、代码评审确认），支持 Approve/Reject |

### 3.2 飞书集成

| 模块 | 实现内容 |
|------|----------|
| **身份认证** | OAuth 2.0 SSO 登录，open_id 全链路治理 |
| **用户管理** | 批量获取用户信息、职位、部门 |
| **部门管理** | 批量获取部门层级、主管、成员信息 |
| **消息推送** | text/post/interactive 多类型消息，支持卡片交互 |
| **文档集成** | docx/wiki 双格式支持，Block 结构读写 |
| **多维表格** | App 创建、数据表 CRUD、记录批量写入 |

### 3.3 AI Function Calling

集成服务（api-ts）提供 14 个 AI 可调用工具：

| 工具 | 功能 |
|------|------|
| `extractContentFromUrls` | 自动识别文本中的飞书文档链接并提取内容 |
| `getDocumentContent` | 根据文档 ID 获取纯文本内容 |
| `createFeishuDocument` | 创建飞书文档 |
| `createFeishuDocumentBlocks` | 批量写入文档内容（支持 block_type 映射） |
| `getAllDocuments` | 获取云盘文件列表 |
| `createFileFolder` | 创建云盘文件夹 |
| `getDepartmentChildren` | 获取子部门列表（支持递归） |
| `createBlock` / `getBlock` / `updateBlock` | 文档块 CRUD |
| `batchUpdateBlocks` / `deleteBlock` | 批量更新/删除 |

---

## 四、功能清单

### 4.1 功能一：Pipeline 引擎（Must-have）

#### 4.1.1 核心能力

| 功能 | 说明 | 状态 |
|------|------|------|
| 阶段定义、排序与依赖管理 | 支持 Stage 定义、排序、依赖管理 | ✅ |
| Agent 编排与执行 | 每个阶段绑定一个或多个 AI Agent | ✅ |
| 阶段间数据流转 | 上一阶段输出作为下一阶段输入 | ✅ |
| Pipeline 生命周期 | 启动、暂停、恢复、终止全流程管理 | ✅ |
| LLM Provider 可配置 | 支持 DeepSeek/OpenAI 等多提供商 | ✅ |
| Human-in-the-Loop | 2 个人工检查点（方案审批、代码评审） | ✅ |
| Approve/Reject 决策 | 支持继续或回退重做 | ✅ |
| Agent 角色定义 | System Prompt + 输入输出契约 | ✅ |
| 代码库上下文感知 | 支持目录/文件路径提供上下文 | ✅ |

#### 4.1.2 API-First 架构

| 功能 | 说明 | 状态 |
|------|------|------|
| Pipeline CRUD | 创建、读取、更新、删除 Pipeline | ✅ |
| Pipeline 执行触发 | `POST /api/pipelines/:id/execute` | ✅ |
| Pipeline 状态查询 | 获取执行状态和进度 | ✅ |
| 检查点审批/拒绝 | Approve/Reject 操作 API | ✅ |
| API 文档 | docs/ 目录下有 API 说明文档 | ✅ |
| **端到端完整演示** | 需求输入 → 全部阶段 → 产出代码变更 | ⬜ |

#### 4.1.3 检查点 UI

| 功能 | 说明 | 状态 |
|------|------|------|
| 检查点列表展示 | 显示所有待审批检查点 | ✅ |
| 检查点详情 | 展示阶段产出物供决策 | ⬜ |
| Approve/Reject 操作 | 审批/拒绝按钮交互 | ⬜ |
| 审批历史记录 | 展示审核流程 | ⬜ |

---

### 4.2 功能二：前端 UI（Must-have）

| 功能 | 说明 | 状态 |
|------|------|------|
| TS+React SPA | 官网风格前端应用 | ✅ |
| 首页与功能页 | 包含 Home/NewRequirement/Session 等页面 | ✅ |
| **悬浮对话框注入** | 页面内注入悬浮对话框 | ⬜ |
| **元素圈选功能** | 支持圈选页面元素 | ⬜ |
| **圈选+对话修改** | 圈选元素后通过对话下达修改指令 | ⬜ |
| **热更新预览** | 修改后实时预览效果 | ⬜ |
| **自动创建 MR** | 自动创建 Merge Request 并生成摘要 | ⬜ |

---

### 4.3 飞书生态集成

| 功能 | 说明 | 状态 |
|------|------|------|
| OAuth SSO 登录 | OAuth 2.0 SSO 登录，open_id 全链路 | ✅ |
| 用户/部门批量获取 | 批量获取用户信息、职位、部门 | ✅ |
| 消息推送 | text/post/interactive 多类型消息 | ✅ |
| 文档读写 | docx/wiki 双格式支持，Block 结构读写 | ✅ |
| 多维表格 CRUD | App 创建、数据表 CRUD、记录写入 | ✅ |
| AI Function Calling | 14 个可调用工具 | ✅ |

---

### 4.4 前端页面详情s

| 页面 | 已实现功能 | 状态 |
|------|-----------|------|
| Home | 8 个快捷提示 · 创建会话（`/api/sessions`） | ✅ |
| NewRequirement | 部门递归获取 · Leader 识别 · 文档选择 · AI 生成文档（SSE）· 消息通知 | ✅ |
| Session | Ant Design X 会话 · SSE 流式 · 乐观渲染 · Markdown 渲染 | ✅ |
| Debug | OAuth Token · 消息发送 · 文档创建 · AI 生成代码 | ✅ |
| Workflows | AntV X6 编辑器（节点增删改、缩放、右键菜单、小地图） | ⬜ |
| Approvals | 检查点审批 | ⬜ |
| Monitoring | Pipeline 监控 | ⬜ |
| Delivery | 测试报告 | ⬜ |

---

### 4.5 加分项（Good-to-have）

| 功能 | 说明 | 状态 |
|------|------|------|
| 多 Agent 协作 | 同阶段多 Agent 并行协商工作 | ⬜ |
| 页面悬浮对话框与圈选 | 注入悬浮对话框，支持元素圈选和对话修改 | ⬜ |
| 可观测性面板 | Pipeline 运行状态实时可视化 | ⬜ |
| 自动回归修复 | Agent 自动修复评审问题并重新提交 | ⬜ |
| 代码库语义索引 | 语义检索提升方案设计和代码生成精准度 | ⬜ |
| Pipeline 模板库 | 预定义 Bug 修复/新功能/重构等模板 | ⬜ |
| 完整 Git 集成 | 自动创建分支、提交代码、发起 MR/PR | ⬜ |
| 测试自动执行 | 自动执行生成的测试用例并返回报告 | ⬜ |

---

## 五、快速开始

### 5.1 环境要求

- Go 1.24+
- Node.js 20+
- pnpm 9+

### 5.2 本地开发

```bash
# 1. 安装依赖
pnpm install

# 2. 配置环境变量
cp apps/api-go/config/config.yaml.example apps/api-go/config/config.yaml
cp apps/api-ts/.env.example apps/api-ts/.env
# 编辑配置文件，填入飞书应用凭证

# 3. 启动 Go 后端（端口 8080）
cd apps/api-go
go run main.go serve

# 4. 启动 TypeScript 中间层（端口 3001，新终端）
pnpm run dev:api-ts

# 5. 启动前端（端口 5173，新终端）
pnpm run dev
```

访问 http://localhost:5173 打开。

### 5.3 Docker 部署

```bash
# 一键启动
docker-compose up -d

# 访问 http://localhost（前端 + 反向代理）
```

---

## 六、API 概览

### 6.1 Go 后端（api-go）— 端口 8080

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `POST /api/auth/feishu` | 飞书 OAuth 登录 |
| `POST /api/pipelines` | 创建 Pipeline |
| `GET/POST /api/pipelines/:id/execute` | 执行 Pipeline |
| `POST /api/pipelines/:id/stages/:sid/approve` | 审批检查点 |
| `POST /api/pipelines/:id/stages/:sid/reject` | 拒绝检查点 |
| `GET /api/sessions` | 会话列表 |
| `POST /api/sessions` | 创建会话 |

### 6.2 TypeScript 中间层（api-ts）— 端口 3001

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `GET /api/health2` | 健康检查（ts） |
| `GET /api/feishu/oauth-url` | 获取 OAuth URL |
| `GET /api/feishu/callback` | OAuth 回调 |
| `POST /api/feishu/send-message` | 发送消息 |
| `POST /api/feishu/create-document` | 创建文档 |
| `GET /api/feishu/document-content` | 获取文档内容 |
| `POST /api/feishu/department-children` | 获取子部门 |
| `POST /api/feishu/batch-departments` | 批量获取部门 |
| `POST /api/feishu/batch-user-names` | 批量获取用户名 |
| `POST /api/bitable/create-app` | 创建多维表格 |
| `POST /api/ai/chat` | AI 对话 |
| `POST /api/ai/chat/stream` | AI 流式对话 |
| `POST /api/ai/get-document-content` | AI 获取文档 |

---

## 七、License

Apache License 2.0