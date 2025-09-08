import { Outlet, useRouteContext } from '@tanstack/react-router';
import { RouterContext } from '../router';
import HamburgerMenu from './HamburgerMenu';
import './Root.css';

export default function Root() {
  const { onLogout } = useRouteContext({ from: '__root__' }) as RouterContext;

  return (
    <div className="app-layout">
      <header className="app-header">
        <div className="app-header-content">
          <h1 className="app-title">Calorie Tracker</h1>
          <HamburgerMenu onLogout={onLogout} />
        </div>
      </header>
      
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}