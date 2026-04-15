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

export default App
