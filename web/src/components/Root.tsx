import { Outlet, useRouteContext } from '@tanstack/react-router';
import { RouterContext } from '../router';
import HamburgerMenu from './HamburgerMenu';
import LanguageSwitcher from './LanguageSwitcher';
// Root.css imports removed - using Tailwind CSS

export default function Root() {
  const { onLogout } = useRouteContext({ from: '__root__' }) as RouterContext;

  return (
    <div className="min-h-screen flex flex-col w-full">
      <header className="bg-white border-b border-gray-200 shadow-sm sticky top-0 z-30">
        <div className="flex justify-between items-center px-5 py-2 max-w-6xl mx-auto md:px-6 md:py-2.5 lg:px-8 lg:py-3">
          <span className="text-2xl font-bold text-gray-800 select-none leading-none">C-Tracker</span>
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