import { profileService } from '@/api/profile';
import { useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { format } from 'date-fns';
import { Calendar, Flame, Target, TrendingDown, TrendingUp } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export default function WeightGoalCard() {
  const { t } = useTranslation();

  const { data: goalProgress, isLoading } = useQuery({
    queryKey: ['weightGoalProgress'],
    queryFn: profileService.getWeightGoalProgress,
  });

  // Don't render anything if no goal is set or still loading
  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-4 mb-4 animate-pulse">
        <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
        <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
        <div className="h-8 bg-gray-200 rounded w-full"></div>
      </div>
    );
  }

  if (!goalProgress) {
    return null; // No goal set, don't show the card
  }

  const progressColor = goalProgress.progress_percent >= 100
    ? 'bg-green-500'
    : goalProgress.progress_percent >= 75
      ? 'bg-emerald-500'
      : goalProgress.progress_percent >= 50
        ? 'bg-blue-500'
        : goalProgress.progress_percent >= 25
          ? 'bg-yellow-500'
          : 'bg-orange-500';

  return (
    <div className="bg-white rounded-lg shadow-md p-4 mb-4">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Target className="w-5 h-5 text-green-600" />
          <h3 className="font-semibold text-gray-800">{t('weightGoal.progressTitle')}</h3>
        </div>
        <Link
          to="/profile"
          className="text-sm text-blue-600 hover:text-blue-800 hover:underline"
        >
          {t('weightGoal.manage')}
        </Link>
      </div>

      {/* Target and Current Weight */}
      <div className="flex items-center justify-between text-sm text-gray-600 mb-2">
        <span>
          {t('weightGoal.target')}: <strong>{goalProgress.target_weight} kg</strong>
        </span>
        <span>
          {t('weightGoal.current')}: <strong>{goalProgress.current_weight.toFixed(1)} kg</strong>
        </span>
      </div>

      {/* Progress Bar */}
      <div className="mb-3">
        <div className="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
          <div
            className={`${progressColor} h-4 rounded-full transition-all duration-500 flex items-center justify-end pr-2`}
            style={{ width: `${Math.min(100, Math.max(5, goalProgress.progress_percent))}%` }}
          >
            {goalProgress.progress_percent >= 15 && (
              <span className="text-xs text-white font-medium">
                {goalProgress.progress_percent.toFixed(0)}%
              </span>
            )}
          </div>
        </div>
        {goalProgress.progress_percent < 15 && (
          <div className="text-right text-xs text-gray-600 mt-0.5">
            {goalProgress.progress_percent.toFixed(0)}%
          </div>
        )}
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-3 gap-2 text-center mb-3">
        <div className="p-2 bg-gray-50 rounded">
          <p className="text-xs text-gray-500">{t('weightGoal.started')}</p>
          <p className="font-semibold text-gray-800">{goalProgress.initial_weight_at_goal.toFixed(1)}</p>
        </div>
        <div className="p-2 bg-gray-50 rounded">
          <p className="text-xs text-gray-500">
            {goalProgress.is_gaining ? t('weightGoal.gained') : t('weightGoal.lost')}
          </p>
          <p className={`font-semibold ${goalProgress.weight_lost > 0 ? 'text-green-600' : 'text-blue-600'}`}>
            {goalProgress.is_gaining ? (
              <TrendingUp className="inline w-3 h-3 mr-1" />
            ) : (
              <TrendingDown className="inline w-3 h-3 mr-1" />
            )}
            {Math.abs(goalProgress.weight_lost).toFixed(1)}
          </p>
        </div>
        <div className="p-2 bg-gray-50 rounded">
          <p className="text-xs text-gray-500">{t('weightGoal.remaining')}</p>
          <p className="font-semibold text-orange-600">{Math.abs(goalProgress.weight_to_go).toFixed(1)}</p>
        </div>
      </div>

      {/* Info Row */}
      <div className="flex flex-wrap gap-3 text-xs text-gray-500">
        {goalProgress.target_date && goalProgress.days_remaining !== undefined && (
          <div className="flex items-center gap-1">
            <Calendar size={12} />
            <span>{t('weightGoal.daysRemaining', { days: goalProgress.days_remaining })}</span>
          </div>
        )}
        {goalProgress.daily_deficit_needed && (
          <div className="flex items-center gap-1">
            <Flame size={12} />
            <span>{t('weightGoal.dailyDeficit', { kcal: Math.round(goalProgress.daily_deficit_needed) })}</span>
          </div>
        )}
        {goalProgress.estimated_completion && !goalProgress.target_date && (
          <div className="flex items-center gap-1">
            <Calendar size={12} />
            <span>
              {t('weightGoal.estimatedShort', {
                date: format(new Date(goalProgress.estimated_completion), 'MMM d'),
              })}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
