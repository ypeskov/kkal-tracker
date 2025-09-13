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
    <header className="flex justify-between items-center mb-6 p-6 bg-white rounded-lg shadow-sm">
      <div>
        <p className="text-gray-800 font-medium text-lg">{t('auth.welcome')}, {user?.email}!</p>
      </div>
    </header>
  );
}