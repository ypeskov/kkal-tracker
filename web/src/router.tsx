import { createRouter, createRoute, createRootRouteWithContext } from '@tanstack/react-router';
import Root from './components/Root';
import DashboardPage from './pages/DashboardPage';
import FoodList from './pages/FoodList';
import Report from './pages/Report';
import Settings from './pages/Settings';

export interface User {
  id: number;
  email: string;
}

export interface RouterContext {
  user: User | undefined;
  onLogout: (() => void) | undefined;
}

const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: Root,
});

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
});

const foodListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/food',
  component: FoodList,
});

const reportRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/reports',
  component: Report,
});

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: Settings,
});

const routeTree = rootRoute.addChildren([
  dashboardRoute,
  foodListRoute,
  reportRoute,
  settingsRoute,
]);

export const router = createRouter({
  routeTree,
  context: undefined! as RouterContext,
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}