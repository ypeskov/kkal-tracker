import { ProfileData } from '@/types/profile';
import { useTranslation } from 'react-i18next';

export default function DashboardHeader({ user }: { user?: ProfileData }) {
  const { t } = useTranslation();

  return (
    <header className="flex justify-between items-center mb-2 p-2 bg-white rounded-lg shadow-sm">
      <span className="text-gray-800 font-medium text-lg block">{t('auth.welcome')}, {user?.first_name} {user?.last_name}!</span>
    </header>
  );
}