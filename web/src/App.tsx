import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import Login from './components/Login';
import { authService } from './api/auth';
// App.css imports removed - using Tailwind CSS
import { router } from './router';
import { RouterProvider } from '@tanstack/react-router';
import i18n from './i18n';

// Helper function to convert backend language codes to i18n format
const convertLanguageCode = (backendCode: string): string => {
  return backendCode.replace('_', '-');
};

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

  const { data: user, error } = useQuery({
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

  if (isInitializing) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-600 text-lg">{t('common.loading')}</div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen bg-gray-100 ${
      !isAuthenticated ? 'flex flex-col items-center justify-center' : ''
    }`}>
      {!isAuthenticated ? (
        <Login onLogin={() => setIsAuthenticated(true)} />
      ) : (
        <RouterProvider router={router} context={{ user: user || undefined, onLogout: handleLogout }} />
      )}
    </div>
  );
}

export default App;