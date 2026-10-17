import { Link, Outlet, useRouteContext } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { RouterContext } from '@/router';
import { prefetchReportPreferences } from '@/hooks/useReportPreferences';
import HamburgerMenu from './HamburgerMenu';
import LanguageSwitcher from './LanguageSwitcher';
// Root.css imports removed - using Tailwind CSS

export default function Root() {
  const { user, onLogout } = useRouteContext({ from: '__root__' }) as RouterContext;
  const queryClient = useQueryClient();
  const userId = user?.id;

  // Load the Reports page settings once per login, so the page usually opens without waiting for them
  useEffect(() => {
    if (userId !== undefined) prefetchReportPreferences(queryClient, userId);
  }, [queryClient, userId]);

  return (
    <div className="min-h-screen flex flex-col w-full">
      <header className="bg-white border-b border-gray-200 shadow-sm sticky top-0 z-30">
        <div className="flex justify-between items-center px-5 py-2 max-w-6xl mx-auto md:px-6 md:py-2.5 lg:px-8 lg:py-3">
          <Link to="/" className="text-xl font-bold text-gray-800 select-none leading-none hover:text-gray-600 transition-colors">
            C-Tracker
          </Link>
          <div className="flex items-center gap-3">
            <LanguageSwitcher />
            <HamburgerMenu onLogout={onLogout} />
          </div>
        </div>
      </header>

      <main className="flex-1 p-0 bg-gray-50 md:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}