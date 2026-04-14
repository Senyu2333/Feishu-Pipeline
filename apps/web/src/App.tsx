import { useEffect, useMemo, useState, type FormEvent } from 'react'
import {
  sessionStatusLabel,
  taskStatusLabel,
  type SessionDetailDTO,
  type SessionSummaryDTO,
  type TaskDTO,
  type TaskStatus,
  type UserDTO,
} from 'shared'

type ApiEnvelope<T> = {
  data: T
  error?: string
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

const TASK_STATUS_OPTIONS: TaskStatus[] = ['todo', 'in_progress', 'testing', 'done']

function App() {
  const [user, setUser] = useState<UserDTO | null>(null)
  const [sessions, setSessions] = useState<SessionSummaryDTO[]>([])
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null)
  const [selectedSession, setSelectedSession] = useState<SessionDetailDTO | null>(null)
  const [title, setTitle] = useState('')
  const [draftPrompt, setDraftPrompt] = useState('')
  const [chatInput, setChatInput] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const selectedTasks = selectedSession?.tasks ?? []

  const publishDisabled = useMemo(() => {
    return !selectedSession || selectedSession.session.status !== 'draft'
  }, [selectedSession])

  useEffect(() => {
    void bootstrap()
  }, [])

  useEffect(() => {
    if (!selectedSessionId) {
      return
    }

    const timer = window.setInterval(() => {
      void loadSession(selectedSessionId, false)
      void loadSessions(false)
    }, 5000)

    return () => window.clearInterval(timer)
  }, [selectedSessionId])

  async function bootstrap() {
    setLoading(true)
    try {
      await Promise.all([loadMe(), loadSessions(true)])
    } catch (err) {
      setError(toErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function loadMe() {
    const response = await apiFetch<UserDTO>('/api/me')
    setUser(response.data)
  }

  async function loadSessions(selectFirst: boolean) {
    const response = await apiFetch<SessionSummaryDTO[]>('/api/sessions')
    setSessions(response.data)

    const nextSelected = selectedSessionId ?? response.data[0]?.id ?? null
    if (selectFirst && nextSelected) {
      setSelectedSessionId(nextSelected)
      await loadSession(nextSelected, true)
    }
  }

  async function loadSession(sessionId: string, updateSelected: boolean) {
    const response = await apiFetch<SessionDetailDTO>(`/api/sessions/${sessionId}`)
    if (updateSelected) {
      setSelectedSessionId(sessionId)
    }
    setSelectedSession(response.data)
  }

  async function handleCreateSession(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!title.trim() || !draftPrompt.trim()) {
      setError('请填写会话标题和需求草稿。')
      return
    }

    setLoading(true)
    setError(null)
    try {
      const response = await apiFetch<SessionDetailDTO>('/api/sessions', {
        method: 'POST',
        body: JSON.stringify({
          title: title.trim(),
          prompt: draftPrompt.trim(),
        }),
      })
      setTitle('')
      setDraftPrompt('')
      setSelectedSessionId(response.data.session.id)
      setSelectedSession(response.data)
      await loadSessions(false)
    } catch (err) {
      setError(toErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function handleSendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!selectedSessionId || !chatInput.trim()) {
      return
    }

    setLoading(true)
    setError(null)
    try {
      const response = await apiFetch<SessionDetailDTO>(`/api/sessions/${selectedSessionId}/messages`, {
        method: 'POST',
        body: JSON.stringify({ content: chatInput.trim() }),
      })
      setChatInput('')
      setSelectedSession(response.data)
      await loadSessions(false)
    } catch (err) {
      setError(toErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function handlePublish() {
    if (!selectedSessionId) {
      return
    }

    setLoading(true)
    setError(null)
    try {
      await apiFetch(`/api/sessions/${selectedSessionId}/publish`, {
        method: 'POST',
        body: JSON.stringify({}),
      })
      await Promise.all([loadSession(selectedSessionId, false), loadSessions(false)])
    } catch (err) {
      setError(toErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  async function handleTaskStatusChange(taskId: string, status: TaskStatus) {
    setError(null)
    try {
      await apiFetch<TaskDTO>(`/api/tasks/${taskId}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      })
      if (selectedSessionId) {
        await Promise.all([loadSession(selectedSessionId, false), loadSessions(false)])
      }
    } catch (err) {
      setError(toErrorMessage(err))
    }
  }

  function handleLogin() {
    window.location.href = `${API_BASE_URL}/api/auth/feishu/login`
  }

  return (
    <div className="workspace-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">AI 驱动需求交付流程引擎</p>
          <h1>Feishu Pipeline 工作台</h1>
        </div>
        <div className="topbar-actions">
          <div className="user-card">
            <strong>{user?.name ?? '未登录用户'}</strong>
            <span>{user ? `${user.role} · ${user.departments.join(' / ')}` : '请先登录飞书或使用本地演示身份'}</span>
          </div>
          <button className="primary ghost" onClick={handleLogin} type="button">
            飞书登录
          </button>
          <button className="primary" disabled={publishDisabled || loading} onClick={handlePublish} type="button">
            发布需求
          </button>
        </div>
      </header>

      {error ? <div className="banner error">{error}</div> : null}
      {loading ? <div className="banner info">正在同步数据，请稍候...</div> : null}

      <main className="layout-grid">
        <aside className="panel sidebar">
          <section className="panel-section">
            <div className="section-title">
              <h2>新建需求会话</h2>
              <span>{sessions.length} 个会话</span>
            </div>
            <form className="stack" onSubmit={handleCreateSession}>
              <input
                placeholder="例如：需求交付引擎 MVP"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
              />
              <textarea
                placeholder="输入需求背景、目标、范围和验收标准..."
                rows={6}
                value={draftPrompt}
                onChange={(event) => setDraftPrompt(event.target.value)}
              />
              <button className="primary" type="submit">
                创建草稿
              </button>
            </form>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <h2>需求会话</h2>
            </div>
            <div className="session-list">
              {sessions.map((session) => (
                <button
                  key={session.id}
                  className={`session-item ${session.id === selectedSessionId ? 'active' : ''}`}
                  type="button"
                  onClick={() => void loadSession(session.id, true)}
                >
                  <div className="session-item-top">
                    <strong>{session.title}</strong>
                    <span className="status-pill">{sessionStatusLabel(session.status)}</span>
                  </div>
                  <p>{session.summary || '暂无摘要'}</p>
                  <div className="session-meta">
                    <span>{session.ownerName}</span>
                    <span>{session.messageCount} 条消息</span>
                  </div>
                </button>
              ))}
            </div>
          </section>
        </aside>

        <section className="panel chat-panel">
          <div className="section-title">
            <div>
              <h2>{selectedSession?.session.title ?? '请选择需求会话'}</h2>
              <span>{selectedSession ? sessionStatusLabel(selectedSession.session.status) : '草稿'}</span>
            </div>
          </div>

          <div className="message-list">
            {selectedSession?.messages.map((message) => (
              <article key={message.id} className={`message-card ${message.role}`}>
                <header>
                  <strong>{message.role === 'user' ? user?.name ?? '我' : 'AI 助手'}</strong>
                  <span>{new Date(message.createdAt).toLocaleString()}</span>
                </header>
                <p>{message.content}</p>
              </article>
            ))}
            {!selectedSession ? <div className="empty-state">左侧创建或选择一个需求会话开始协作。</div> : null}
          </div>

          <form className="chat-composer" onSubmit={handleSendMessage}>
            <textarea
              placeholder="继续补充需求细节，或在发布后对任务进行澄清..."
              rows={4}
              value={chatInput}
              onChange={(event) => setChatInput(event.target.value)}
              disabled={!selectedSession}
            />
            <button className="primary" type="submit" disabled={!selectedSession || loading}>
              发送消息
            </button>
          </form>
        </section>

        <aside className="panel detail-panel">
          <section className="panel-section">
            <div className="section-title">
              <h2>需求摘要</h2>
            </div>
            {selectedSession?.requirement ? (
              <div className="stack">
                <p>{selectedSession.requirement.summary}</p>
                <div className="detail-meta">
                  <span>发布状态：{sessionStatusLabel(selectedSession.requirement.status)}</span>
                  <span>
                    发布时间：
                    {selectedSession.requirement.publishedAt
                      ? new Date(selectedSession.requirement.publishedAt).toLocaleString()
                      : '处理中'}
                  </span>
                </div>
                <div className="tag-list">
                  {selectedSession.requirement.referencedKnowledge.map((item) => (
                    <span className="knowledge-chip" key={item}>
                      {item}
                    </span>
                  ))}
                </div>
              </div>
            ) : (
              <div className="empty-state">当前还是草稿会话，发布后会在这里看到正式摘要和引用知识。</div>
            )}
          </section>

          <section className="panel-section">
            <div className="section-title">
              <h2>任务拆解</h2>
              <span>{selectedTasks.length} 项</span>
            </div>
            <div className="task-list">
              {selectedTasks.map((task) => (
                <TaskCard key={task.id} task={task} onStatusChange={handleTaskStatusChange} />
              ))}
              {!selectedTasks.length ? <div className="empty-state">发布需求后将自动生成前后端任务和交付链接。</div> : null}
            </div>
          </section>
        </aside>
      </main>
    </div>
  )
}

function TaskCard(props: { task: TaskDTO; onStatusChange: (taskId: string, status: TaskStatus) => Promise<void> }) {
  const { task, onStatusChange } = props

  return (
    <article className="task-card">
      <header>
        <div>
          <strong>{task.title}</strong>
          <span>{task.assigneeName}</span>
        </div>
        <span className="status-pill">{taskStatusLabel(task.status)}</span>
      </header>
      <p>{task.description}</p>
      <div className="task-links">
        {task.docURL ? (
          <a href={task.docURL} target="_blank" rel="noreferrer">
            任务文档
          </a>
        ) : null}
        {task.bitableRecordURL ? (
          <a href={task.bitableRecordURL} target="_blank" rel="noreferrer">
            多维表格
          </a>
        ) : null}
      </div>
      <label className="task-status-editor">
        <span>状态</span>
        <select value={task.status} onChange={(event) => void onStatusChange(task.id, event.target.value as TaskStatus)}>
          {TASK_STATUS_OPTIONS.map((status) => (
            <option key={status} value={status}>
              {taskStatusLabel(status)}
            </option>
          ))}
        </select>
      </label>
    </article>
  )
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<ApiEnvelope<T>> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  const payload = (await response.json()) as ApiEnvelope<T>
  if (!response.ok) {
    throw new Error(payload.error ?? '请求失败')
  }
  return payload
}

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message
  }
  return '发生未知错误'
}

export default App
