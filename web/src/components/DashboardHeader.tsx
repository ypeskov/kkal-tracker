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
    <header className="flex justify-between items-center mb-2 p-2 bg-white rounded-lg shadow-sm">
      <span className="text-gray-800 font-medium text-lg block">{t('auth.welcome')}, {user?.email}!</span>
    </header>
  );
}