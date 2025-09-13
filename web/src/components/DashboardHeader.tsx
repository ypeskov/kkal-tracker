import { useTranslation } from 'react-i18next';

interface User {
  id: number;
  email: string;
}

interface DashboardHeaderProps {
  user?: User;
}

export default function DashboardHeader({ user }: DashboardHeaderProps) {
  const { t } = useTranslation();

  return (
    <header className="dashboard-header">
      <div>
        <p>{t('auth.welcome')}, {user?.email}!</p>
      </div>
    </header>
  );
}