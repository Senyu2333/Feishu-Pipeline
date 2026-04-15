# Feishu Pipeline API 接口文档

面向 `apps/api-go` 服务的 HTTP 接口说明。框架为 **Gin**，默认监听端口见配置 `app.port`（常为 `8080`）。

## 通用约定

| 项目 | 说明 |
|------|------|
| Base URL | 由部署环境决定，开发可参考配置 `app.base_url`（如 `http://localhost:8080`） |
| 协议 | HTTP/HTTPS |
| 请求体 | `Content-Type: application/json`（除无 body 的 GET 等） |
| 响应体 | 统一信封结构（见下） |
| 认证 | 登录成功后服务端通过 **HttpOnly Cookie** 下发会话；Cookie 名称为配置 `app.session_cookie_name`（默认 `feishu_pipeline_session`）。需登录的接口必须携带有效会话 Cookie |

### 响应信封

```json
{
  "data": {}
}
```

错误时：

```json
{
  "error": "错误说明文字"
}
```

### CORS

中间件允许携带凭证（`Access-Control-Allow-Credentials: true`），浏览器跨域时需与前端 Origin 策略一致。

### HTTP 状态码（常见）

| 状态码 | 含义 |
|--------|------|
| 200 | 成功 |
| 201 | 已创建 |
| 202 | 已接受（异步） |
| 400 | 请求参数或业务校验失败 |
| 401 | 未登录或会话无效 |
| 403 | 无权限（如非管理员访问管理接口） |
| 404 | 资源不存在 |
| 502 | 上游服务失败（如飞书登录） |

---

## 枚举与类型参考

### 用户角色 `role`

`product` | `frontend` | `backend` | `admin`

### 会话状态 `SessionStatus`

`draft` | `published` | `in_delivery` | `testing` | `done` | `archived`

### 消息角色 `MessageRole`

`user` | `assistant` | `system`

### 任务类型 `TaskType`

`frontend` | `backend` | `shared`

### 任务优先级 `TaskPriority`

`high` | `medium` | `low`

### 任务状态 `TaskStatus`

`todo` | `in_progress` | `testing` | `done`

---

## 公开接口（无需登录）

### GET `/api/health`

健康检查。

**响应 `data` 示例字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | 如 `ok` |
| service | string | 服务名 |
| version | string | 版本 |
| now | string | UTC 时间 RFC3339 |

---

### GET `/api/auth/feishu/config`

获取飞书 SSO 前端配置。

**响应 `data`**

| 字段 | 类型 | 说明 |
|------|------|------|
| enabled | bool | 是否启用飞书登录 |
| appId | string | 飞书应用 App ID（未配置时可能为空） |

---

### POST `/api/auth/feishu/sso/login`

使用飞书 OAuth 返回的授权码完成登录。

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 飞书授权码 |

**响应 `data`**

| 字段 | 类型 | 说明 |
|------|------|------|
| user | object | 当前用户，字段见「用户对象」 |

成功时同时 **Set-Cookie** 写入会话。

**常见错误**：`502` 飞书或登录流程失败。

---

### POST `/api/auth/logout`

登出，清除服务端会话并清除 Cookie。

**请求体**：无

**响应 `data`**

```json
{ "status": "logged_out" }
```

---

## 需登录接口（Cookie）

以下路径均需有效会话；未登录返回 **401**，`error` 为认证相关说明。

### GET `/api/me`

当前登录用户信息。

**响应 `data`：用户对象**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |
| feishuOpenID | string | 飞书 Open ID（可选） |
| name | string | 姓名 |
| email | string | 邮箱（可选） |
| role | string | 见「用户角色」 |
| departments | string[] | 部门 |

---

### GET `/api/sessions`

当前用户的会话列表。

**响应 `data`**：`SessionSummaryResponse[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 会话 ID |
| title | string | 标题 |
| summary | string | 摘要 |
| status | string | 见「会话状态」 |
| ownerName | string | 负责人姓名 |
| messageCount | int | 消息条数 |
| updatedAt | string | ISO8601 时间 |

---

### POST `/api/sessions`

创建会话。

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题 |
| prompt | string | 是 | 初始提示/需求描述 |

**响应**：**201**，`data` 为会话详情（结构同「GET 会话详情」）。

---

### GET `/api/sessions/:sessionID`

会话详情。

**路径参数**：`sessionID` 会话 ID。

**响应 `data`：SessionDetailResponse**

| 字段 | 类型 | 说明 |
|------|------|------|
| session | object | 摘要，同列表项结构 |
| messages | array | 消息列表 |
| requirement | object \| null | 需求信息，无则省略或 null |
| tasks | array | 任务列表 |

**消息项 MessageResponse**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 消息 ID |
| sessionId | string | 会话 ID |
| role | string | 见「消息角色」 |
| content | string | 内容 |
| createdAt | string | 创建时间 |

**需求项 RequirementResponse**（存在时）

| 字段 | 类型 | 说明 |
|------|------|------|
| sessionId | string | 会话 ID |
| requirementId | string | 需求 ID |
| title | string | 标题 |
| summary | string | 摘要 |
| status | string | 会话状态 |
| publishedAt | string \| null | 发布时间 |
| deliverySummary | string | 交付说明（可选） |
| referencedKnowledge | string[] | 引用知识 |

**任务项**：结构见「GET 任务详情」。

---

### POST `/api/sessions/:sessionID/messages`

追加一条用户消息。

**路径参数**：`sessionID`

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 消息正文 |

**响应**：**201**，`data` 为更新后的完整会话详情。

---

### POST `/api/sessions/:sessionID/publish`

发布会话（后台异步处理需求拆解、任务与飞书分发等）。

**路径参数**：`sessionID`

**请求体**：无

**响应**：**202**，`data` 示例：

```json
{
  "status": "accepted",
  "message": "需求已受理，后台正在生成任务和飞书分发结果。"
}
```

---

### GET `/api/tasks/:taskID`

任务详情。

**路径参数**：`taskID`

**响应 `data`：TaskResponse**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 任务 ID |
| sessionId | string | 所属会话 |
| title | string | 标题 |
| description | string | 描述 |
| type | string | 任务类型 |
| status | string | 任务状态 |
| priority | string | 优先级 |
| estimateDays | int | 预估人天 |
| assigneeName | string | 指派人姓名 |
| assigneeRole | string | 指派角色 |
| assigneeId | string | 指派 ID（可选） |
| assigneeIdType | string | ID 类型（可选） |
| plannedStartAt | string \| null | 计划开始 |
| plannedEndAt | string \| null | 计划结束 |
| notifyContent | string | 通知文案（可选） |
| docURL | string | 文档链接（可选） |
| bitableRecordURL | string | 多维表格记录（可选） |
| acceptanceCriteria | string[] | 验收标准 |
| risks | string[] | 风险 |
| createdAt | string | 创建时间 |
| updatedAt | string | 更新时间 |

---

### PATCH `/api/tasks/:taskID/status`

更新任务状态。

**路径参数**：`taskID`

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 是 | 见「任务状态」 |

**响应**：**200**，`data` 为更新后的任务对象。

---

## 管理员接口（需登录且 `role === admin`）

非管理员返回 **403**，`error` 含权限说明。

### POST `/api/admin/role-mappings`

创建或保存角色映射规则。

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 规则名称 |
| keyword | string | 是 | 匹配关键词 |
| role | string | 是 | 用户角色 |
| departments | string[] | 否 | 部门列表 |

**响应**：**201**

```json
{ "status": "saved" }
```

---

### GET `/api/admin/role-owners`

角色负责人列表。

**响应 `data`**：`RoleOwnerResponse[]`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 记录 ID |
| role | string | 角色 |
| ownerName | string | 负责人姓名 |
| feishuId | string | 飞书 ID（可选） |
| feishuIdType | string | ID 类型（可选） |
| enabled | bool | 是否启用 |

---

### POST `/api/admin/role-owners`

保存角色负责人。

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | 是 | 角色 |
| ownerName | string | 是 | 负责人姓名 |
| feishuId | string | 否 | 飞书 ID |
| feishuIdType | string | 否 | ID 类型 |
| enabled | bool | 否 | 是否启用 |

**响应**：**201**

```json
{ "status": "saved" }
```

---

### POST `/api/admin/knowledge/sync`

批量同步知识库来源。

**请求体**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sources | array | 是 | 知识条目列表 |

**sources[] 每项**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题 |
| content | string | 是 | 正文 |

**响应**：**201**

| 字段 | 类型 | 说明 |
|------|------|------|
| count | int | 写入条数 |

---

## 相关文件

| 说明 | 路径 |
|------|------|
| 路由定义 | `apps/api-go/internal/router/router.go` |
| Postman 集合（可导入） | `docs/postman/Feishu-Pipeline-API.postman_collection.json` |
| 应用配置示例 | `apps/api-go/config/config.yaml` |

---

## 修订说明

文档随代码演进，若接口行为与本文不一致，以 `router.go` 与各 `*_controller.go` 实现为准。
