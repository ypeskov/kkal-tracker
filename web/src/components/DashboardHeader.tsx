import { useTranslation } from 'react-i18next';
import LanguageSwitcher from './LanguageSwitcher';

interface User {
  id: number;
  email: string;
}

interface DashboardHeaderProps {
  user?: User;
  onLogout?: () => void;
}

export default function DashboardHeader({ user, onLogout }: DashboardHeaderProps) {
  const { t } = useTranslation();

  return (
    <header className="dashboard-header">
      <div>
        <h1>{t('dashboard.title')}</h1>
        <p>{t('auth.welcome')}, {user?.email}!</p>
      </div>
      <div className="header-controls">
        <LanguageSwitcher />
        <button 
          onClick={onLogout}
          className="btn"
          style={{ backgroundColor: '#dc3545' }}
        >
          {t('auth.logout')}
        </button>
      </div>
    </header>
  );
}