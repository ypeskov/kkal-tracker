import { ProfileData } from '@/types/profile';
import { createRootRouteWithContext, createRoute, createRouter } from '@tanstack/react-router';
import Root from './components/Root';
import ActivationPage from './pages/ActivationPage';
import AIInsights from './pages/AIInsights';
import DashboardPage from './pages/DashboardPage';
import FoodList from './pages/FoodList';
import Profile from './pages/Profile';
import RegisterPage from './pages/RegisterPage';
import Report from './pages/Report';

export interface RouterContext {
  user: ProfileData | undefined;
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

const profileRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/profile',
  component: Profile,
});

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/register',
  component: RegisterPage,
});

const activationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activate/$token',
  component: ActivationPage,
});

const aiInsightsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/ai-insights',
  component: AIInsights,
});

const routeTree = rootRoute.addChildren([
  dashboardRoute,
  foodListRoute,
  reportRoute,
  aiInsightsRoute,
  profileRoute,
  registerRoute,
  activationRoute,
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