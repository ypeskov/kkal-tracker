import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import Login from './components/Login';
import { authService } from './api/auth';
import './App.css';
import { router } from './router';
import { RouterProvider } from '@tanstack/react-router';

function App() {
  const { t } = useTranslation();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isInitializing, setIsInitializing] = useState(true);

  useEffect(() => {
    const token = authService.getToken();
    if (token) {
      setIsAuthenticated(true);
    }
    setIsInitializing(false);
  }, []);

  const { data: user, isLoading, error } = useQuery({
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

  const handleLogout = () => {
    authService.logout();
    setIsAuthenticated(false);
  };

  if (isInitializing || isLoading) {
    return <div>{t('common.loading')}</div>;
  }

  return (
    <div className="app">
      {!isAuthenticated ? (
        <Login onLogin={() => setIsAuthenticated(true)} />
      ) : (
        <RouterProvider router={router} context={{ user, onLogout: handleLogout }} />
      )}
    </div>
  );
}

export default App;