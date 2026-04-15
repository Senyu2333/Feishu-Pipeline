import { createRouter, RouterProvider } from '@tanstack/react-router'
import { createRoute, createRootRoute } from '@tanstack/react-router'
import Home from './pages/Home'
import NewRequirement from './pages/NewRequirement'
import Workflows from './pages/Workflows'
import Monitoring from './pages/Monitoring'
import Approvals from './pages/Approvals'
import Delivery from './pages/Delivery'

const rootRoute = createRootRoute()

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: Home,
})

const newRequirementRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/new-requirement',
  component: NewRequirement,
})

const workflowsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/workflows',
  component: Workflows,
})

const monitoringRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/monitoring',
  component: Monitoring,
})

const approvalsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/approvals',
  component: Approvals,
})

const deliveryRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/delivery',
  component: Delivery,
})

const routeTree = rootRoute.addChildren([indexRoute, newRequirementRoute, workflowsRoute, monitoringRoute, approvalsRoute, deliveryRoute])
const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

function App() {
  return <RouterProvider router={router} />
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
