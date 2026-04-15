export type SystemRole = 'product' | 'frontend' | 'backend' | 'admin'

export type SessionStatus =
  | 'draft'
  | 'published'
  | 'in_delivery'
  | 'testing'
  | 'done'
  | 'archived'

export type MessageRole = 'user' | 'assistant' | 'system'

export type TaskType = 'frontend' | 'backend' | 'shared'

export type TaskStatus = 'todo' | 'in_progress' | 'testing' | 'done'

export interface UserDTO {
  id: string
  feishuOpenID?: string
  name: string
  email?: string
  role: SystemRole
  departments: string[]
}

export interface MessageDTO {
  id: string
  sessionId: string
  role: MessageRole
  content: string
  createdAt: string
}

export interface TaskDTO {
  id: string
  sessionId: string
  title: string
  description: string
  type: TaskType
  status: TaskStatus
  assigneeName: string
  assigneeRole: SystemRole
  docURL?: string
  bitableRecordURL?: string
  acceptanceCriteria: string[]
  risks: string[]
  createdAt: string
  updatedAt: string
}

export interface SessionSummaryDTO {
  id: string
  title: string
  summary: string
  status: SessionStatus
  ownerName: string
  updatedAt: string
  messageCount: number
}

export interface RequirementDetailDTO {
  sessionId: string
  requirementId?: string
  title: string
  summary: string
  status: SessionStatus
  publishedAt?: string
  deliverySummary?: string
  referencedKnowledge: string[]
  tasks: TaskDTO[]
}

export interface SessionDetailDTO {
  session: SessionSummaryDTO
  messages: MessageDTO[]
  requirement?: RequirementDetailDTO
  tasks: TaskDTO[]
}

export interface HealthDTO {
  status: 'ok'
  service: string
  version: string
  now: string
}

export interface ApiEnvelope<T> {
  data: T
  error?: string
}

export interface CreateSessionRequest {
  title: string
  prompt: string
}

export interface CreateMessageRequest {
  content: string
}

export interface PublishSessionRequest {
  force?: boolean
}

export interface UpdateTaskStatusRequest {
  status: TaskStatus
}

export interface RoleMappingRequest {
  name: string
  keyword: string
  role: SystemRole
  departments: string[]
}

export interface KnowledgeSyncRequest {
  sources: Array<{
    title: string
    content: string
  }>
}

export function sessionStatusLabel(status: SessionStatus): string {
  const labels: Record<SessionStatus, string> = {
    draft: '草稿',
    published: '已发布',
    in_delivery: '交付中',
    testing: '测试中',
    done: '已完成',
    archived: '已归档',
  }
  return labels[status]
}

export function taskStatusLabel(status: TaskStatus): string {
  const labels: Record<TaskStatus, string> = {
    todo: '未开发',
    in_progress: '正在开发',
    testing: '已提测',
    done: '完成',
  }
  return labels[status]
}
