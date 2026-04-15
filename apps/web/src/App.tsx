import { createRouter, RouterProvider } from '@tanstack/react-router'
import { createRoute, createRootRoute } from '@tanstack/react-router'
import Home from './pages/Home'
import NewRequirement from './pages/NewRequirement'
import Workflows from './pages/Workflows'
import Monitoring from './pages/Monitoring'
import Approvals from './pages/Approvals'
import Delivery from './pages/Delivery'
import { TaskDTO, TaskStatus, ApiEnvelope, taskStatusLabel } from 'shared'

const API_BASE_URL = '/api'

const TASK_STATUS_OPTIONS: TaskStatus[] = ['todo', 'in_progress', 'testing', 'done']

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
    <article className="bg-white rounded-xl p-4 shadow-sm border border-gray-100">
      <header className="flex items-center justify-between mb-3">
        <div className="flex flex-col gap-1">
          <strong className="text-sm font-semibold text-gray-800">{task.title}</strong>
          <span className="text-xs text-gray-500">{task.assigneeName}</span>
        </div>
        <span className="px-2 py-0.5 text-xs font-medium rounded-md bg-blue-50 text-blue-600">{taskStatusLabel(task.status)}</span>
      </header>
      <p className="text-sm text-gray-600 mb-3">{task.description}</p>
      <div className="flex gap-3 mb-3">
        {task.docURL ? (
          <a href={task.docURL} target="_blank" rel="noreferrer" className="text-xs text-blue-600 hover:underline">
            任务文档
          </a>
        ) : null}
        {task.bitableRecordURL ? (
          <a href={task.bitableRecordURL} target="_blank" rel="noreferrer" className="text-xs text-blue-600 hover:underline">
            多维表格
          </a>
        ) : null}
      </div>
      <label className="flex items-center gap-2 text-sm">
        <span className="text-gray-500">状态</span>
        <select 
          value={task.status} 
          onChange={(event) => void onStatusChange(task.id, event.target.value as TaskStatus)}
          className="flex-1 px-2 py-1 text-sm border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
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
