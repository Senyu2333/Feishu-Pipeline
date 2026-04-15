
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

  const payload = response.headers.get('content-type')?.includes('application/json')
    ? ((await response.json()) as ApiEnvelope<T>)
    : ({ error: await response.text() } as ApiEnvelope<T>)
  if (!response.ok) {
    throw new ApiError(payload.error ?? '请求失败', response.status)
  }
  return payload
}

class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

function toErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message
  }
  return '发生未知错误'
}

export default App
