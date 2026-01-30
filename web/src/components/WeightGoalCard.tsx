import { profileService } from '@/api/profile';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import WeightGoalProgressDisplay from './WeightGoalProgressDisplay';

export default function WeightGoalCard() {
  const { t } = useTranslation();

  const { data: goalProgress, isLoading } = useQuery({
    queryKey: ['weightGoalProgress'],
    queryFn: profileService.getWeightGoalProgress,
  });

  // Loading state
  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-4 mb-4 animate-pulse">
        <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
        <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
        <div className="h-8 bg-gray-200 rounded w-full"></div>
      </div>
    );
  }

  // No goal set
  if (!goalProgress) {
    return null;
  }

  return (
    <div className="mb-4">
      <WeightGoalProgressDisplay
        goalProgress={goalProgress}
        headerAction={
          <Link
            to="/profile"
            className="text-sm text-blue-600 hover:text-blue-800 hover:underline"
          >
            {t('weightGoal.manage')}
          </Link>
        }
      />
    </div>
  );
}
