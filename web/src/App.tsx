import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import Login from './components/Login';
import { authService } from './api/auth';
// App.css imports removed - using Tailwind CSS
import { router } from './router';
import { RouterProvider } from '@tanstack/react-router';
import i18n from './i18n';
import RegisterPage from './pages/RegisterPage';
import ActivationPage from './pages/ActivationPage';

// Helper function to convert backend language codes to i18n format
const convertLanguageCode = (backendCode: string): string => {
  return backendCode.replace('_', '-');
};

function App() {
  const { t } = useTranslation();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isInitializing, setIsInitializing] = useState(true);
  const [currentPath, setCurrentPath] = useState(window.location.pathname);

  // Track pathname changes
  useEffect(() => {
    const handleLocationChange = () => {
      setCurrentPath(window.location.pathname);
    };

    window.addEventListener('popstate', handleLocationChange);

    // Also track programmatic navigation
    const originalPushState = window.history.pushState;
    const originalReplaceState = window.history.replaceState;

    window.history.pushState = function(...args) {
      originalPushState.apply(window.history, args);
      handleLocationChange();
    };

    window.history.replaceState = function(...args) {
      originalReplaceState.apply(window.history, args);
      handleLocationChange();
    };

    return () => {
      window.removeEventListener('popstate', handleLocationChange);
      window.history.pushState = originalPushState;
      window.history.replaceState = originalReplaceState;
    };
  }, []);

  useEffect(() => {
    const token = authService.getToken();
    if (token) {
      setIsAuthenticated(true);
    }
    setIsInitializing(false);
  }, []);

  const { data: user, error, isLoading: isUserLoading } = useQuery({
    queryKey: ['user'],
    queryFn: authService.getCurrentUser,
    retry: false,
    enabled: isAuthenticated,
  });

  useEffect(() => {
    if (error && isAuthenticated) {
      authService.logout();
      setIsAuthenticated(false);
    }
  }, [error, isAuthenticated]);

  // Apply user's language preference when user data is loaded
  useEffect(() => {
    if (user && user.language) {
      const i18nLanguage = convertLanguageCode(user.language);
      if (i18n.language !== i18nLanguage) {
        i18n.changeLanguage(i18nLanguage);
      }
    }
  }, [user]);

  const handleLogout = () => {
    authService.logout();
    setIsAuthenticated(false);
  };

  if (isInitializing || (isAuthenticated && isUserLoading)) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-600 text-lg">{t('common.loading')}</div>
      </div>
    );
  }

  // Check if current path is a public route (register or activate)
  const isPublicRoute = currentPath === '/register' || currentPath.startsWith('/activate/');

  // Render public pages directly without RouterProvider
  if (!isAuthenticated && isPublicRoute) {
    if (currentPath === '/register') {
      return <RegisterPage />;
    }
    if (currentPath.startsWith('/activate/')) {
      const token = currentPath.split('/activate/')[1];
      return <ActivationPage token={token} />;
    }
  }

  // Show login for unauthenticated users on protected routes
  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center">
        <Login onLogin={() => setIsAuthenticated(true)} />
      </div>
    );
  }

  // Show app for authenticated users
  return (
    <div className="min-h-screen bg-gray-100">
      <RouterProvider router={router} context={{ user: user!, onLogout: handleLogout }} />
    </div>
  );
}

export default App;