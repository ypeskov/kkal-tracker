import { profileService } from '@/api/profile';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { format, addDays } from 'date-fns';
import { Target, Trash2, AlertTriangle, TrendingDown, TrendingUp, Calendar, Flame } from 'lucide-react';
import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

interface WeightGoalFormProps {
  currentWeight?: number;
  onSuccess?: () => void;
}

export default function WeightGoalForm({ currentWeight, onSuccess }: WeightGoalFormProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  
  const [targetWeight, setTargetWeight] = useState('');
  const [targetDate, setTargetDate] = useState('');
  const [noDeadline, setNoDeadline] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch current goal progress
  const { data: goalProgress, isLoading } = useQuery({
    queryKey: ['weightGoalProgress'],
    queryFn: profileService.getWeightGoalProgress,
  });

  // Set weight goal mutation
  const setGoalMutation = useMutation({
    mutationFn: profileService.setWeightGoal,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['weightGoalProgress'] });
      queryClient.invalidateQueries({ queryKey: ['profile'] });
      setShowForm(false);
      setError(null);
      onSuccess?.();
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  // Clear weight goal mutation
  const clearGoalMutation = useMutation({
    mutationFn: profileService.clearWeightGoal,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['weightGoalProgress'] });
      queryClient.invalidateQueries({ queryKey: ['profile'] });
      setTargetWeight('');
      setTargetDate('');
      setNoDeadline(true);
      setError(null);
    },
    onError: (err: Error) => {
      setError(err.message);
    },
  });

  // Initialize form with existing goal data
  useEffect(() => {
    if (goalProgress) {
      setTargetWeight(goalProgress.target_weight.toString());
      if (goalProgress.target_date) {
        setTargetDate(format(new Date(goalProgress.target_date), 'yyyy-MM-dd'));
        setNoDeadline(false);
      } else {
        setNoDeadline(true);
      }
    }
  }, [goalProgress]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const weight = parseFloat(targetWeight);
    if (isNaN(weight) || weight < 30 || weight > 300) {
      setError(t('weightGoal.invalidWeight'));
      return;
    }

    setGoalMutation.mutate({
      target_weight: weight,
      target_date: noDeadline ? undefined : targetDate || undefined,
    });
  };

  const handleClearGoal = () => {
    if (confirm(t('weightGoal.confirmClear'))) {
      clearGoalMutation.mutate();
    }
  };

  // Calculate recommendations based on form inputs
  const calculateRecommendations = () => {
    if (!currentWeight || !targetWeight) return null;
    
    const target = parseFloat(targetWeight);
    if (isNaN(target)) return null;

    const weightDiff = Math.abs(currentWeight - target);
    const isLosing = currentWeight > target;

    if (noDeadline) {
      // Estimate at 0.5 kg/week
      const weeksNeeded = weightDiff / 0.5;
      const estimatedDate = addDays(new Date(), Math.ceil(weeksNeeded * 7));
      return {
        dailyDeficit: Math.round(weightDiff * 7700 / (weeksNeeded * 7)),
        estimatedDate,
        isLosing,
        isSafe: true,
      };
    } else if (targetDate) {
      const days = Math.ceil((new Date(targetDate).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24));
      if (days <= 0) return null;
      
      const dailyDeficit = Math.round((weightDiff * 7700) / days);
      const weeklyLoss = weightDiff / (days / 7);
      
      return {
        dailyDeficit,
        days,
        isLosing,
        isSafe: weeklyLoss <= 1,
        weeklyRate: weeklyLoss,
      };
    }
    return null;
  };

  const recommendations = calculateRecommendations();

  if (isLoading) {
    return (
      <div className="p-4 bg-gray-50 rounded-lg animate-pulse">
        <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
        <div className="h-20 bg-gray-200 rounded"></div>
      </div>
    );
  }

  // Display current goal progress
  if (goalProgress && !showForm) {
    return (
      <div className="space-y-4">
        {/* Progress Card */}
        <div className="p-4 bg-gradient-to-r from-green-50 to-emerald-50 rounded-lg border border-green-200">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Target className="w-5 h-5 text-green-600" />
              <h3 className="font-semibold text-gray-800">{t('weightGoal.currentGoal')}</h3>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setShowForm(true)}
                className="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
              >
                {t('common.edit')}
              </button>
              <button
                onClick={handleClearGoal}
                disabled={clearGoalMutation.isPending}
                className="p-1.5 text-red-600 hover:bg-red-50 rounded-md transition-colors"
                title={t('weightGoal.clearGoal')}
              >
                <Trash2 size={16} />
              </button>
            </div>
          </div>

          {/* Progress Bar */}
          <div className="mb-4">
            <div className="flex justify-between text-sm mb-1">
              <span className="text-gray-600">
                {goalProgress.is_gaining ? t('weightGoal.targetGain') : t('weightGoal.targetLoss')}: {goalProgress.target_weight} kg
              </span>
              <span className="font-semibold text-green-600">{goalProgress.progress_percent.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-3">
              <div
                className="bg-green-500 h-3 rounded-full transition-all duration-500"
                style={{ width: `${Math.min(100, goalProgress.progress_percent)}%` }}
              ></div>
            </div>
          </div>

          {/* Stats Grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <div className="text-center p-2 bg-white/50 rounded-lg">
              <p className="text-xs text-gray-500">{t('weightGoal.started')}</p>
              <p className="font-bold text-gray-800">{goalProgress.initial_weight_at_goal.toFixed(1)} kg</p>
            </div>
            <div className="text-center p-2 bg-white/50 rounded-lg">
              <p className="text-xs text-gray-500">{t('weightGoal.current')}</p>
              <p className="font-bold text-gray-800">{goalProgress.current_weight.toFixed(1)} kg</p>
            </div>
            <div className="text-center p-2 bg-white/50 rounded-lg">
              <p className="text-xs text-gray-500">
                {goalProgress.is_gaining ? t('weightGoal.gained') : t('weightGoal.lost')}
              </p>
              <p className={`font-bold ${goalProgress.weight_lost > 0 ? 'text-green-600' : 'text-blue-600'}`}>
                {Math.abs(goalProgress.weight_lost).toFixed(1)} kg
              </p>
            </div>
            <div className="text-center p-2 bg-white/50 rounded-lg">
              <p className="text-xs text-gray-500">{t('weightGoal.remaining')}</p>
              <p className="font-bold text-orange-600">{Math.abs(goalProgress.weight_to_go).toFixed(1)} kg</p>
            </div>
          </div>

          {/* Additional Info */}
          <div className="mt-3 pt-3 border-t border-green-200 flex flex-wrap gap-4 text-sm">
            {goalProgress.target_date && goalProgress.days_remaining !== undefined && (
              <div className="flex items-center gap-1.5 text-gray-600">
                <Calendar size={14} />
                <span>
                  {t('weightGoal.daysRemaining', { days: goalProgress.days_remaining })}
                </span>
              </div>
            )}
            {goalProgress.daily_deficit_needed && (
              <div className="flex items-center gap-1.5 text-gray-600">
                <Flame size={14} />
                <span>
                  {t('weightGoal.dailyDeficit', { kcal: Math.round(goalProgress.daily_deficit_needed) })}
                </span>
              </div>
            )}
            {goalProgress.estimated_completion && !goalProgress.target_date && (
              <div className="flex items-center gap-1.5 text-gray-600">
                <Calendar size={14} />
                <span>
                  {t('weightGoal.estimatedCompletion', {
                    date: format(new Date(goalProgress.estimated_completion), 'PP'),
                  })}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  // Form to set/edit goal
  return (
    <div className="space-y-4">
      {goalProgress && (
        <button
          onClick={() => setShowForm(false)}
          className="text-sm text-gray-500 hover:text-gray-700"
        >
          ← {t('common.back')}
        </button>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Target Weight */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('weightGoal.targetWeight')} (kg) *
            </label>
            <div className="relative">
              <input
                type="number"
                step="0.1"
                min="30"
                max="300"
                value={targetWeight}
                onChange={(e) => setTargetWeight(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                placeholder={t('weightGoal.enterTarget')}
                required
              />
              {currentWeight && (
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400">
                  {t('weightGoal.currentIs', { weight: currentWeight.toFixed(1) })}
                </span>
              )}
            </div>
          </div>

          {/* Target Date */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('weightGoal.targetDate')}
            </label>
            <input
              type="date"
              value={targetDate}
              onChange={(e) => setTargetDate(e.target.value)}
              disabled={noDeadline}
              min={format(addDays(new Date(), 1), 'yyyy-MM-dd')}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500 disabled:bg-gray-100 disabled:text-gray-500"
            />
          </div>
        </div>

        {/* No Deadline Checkbox */}
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="noDeadline"
            checked={noDeadline}
            onChange={(e) => {
              setNoDeadline(e.target.checked);
              if (e.target.checked) {
                setTargetDate('');
              }
            }}
            className="w-4 h-4 text-green-600 border-gray-300 rounded focus:ring-green-500"
          />
          <label htmlFor="noDeadline" className="text-sm text-gray-600">
            {t('weightGoal.noDeadline')}
          </label>
        </div>

        {/* Recommendations */}
        {recommendations && currentWeight && (
          <div className={`p-3 rounded-lg ${recommendations.isSafe ? 'bg-blue-50 border border-blue-200' : 'bg-amber-50 border border-amber-200'}`}>
            <div className="flex items-start gap-2">
              {recommendations.isSafe ? (
                recommendations.isLosing ? (
                  <TrendingDown className="w-5 h-5 text-blue-600 mt-0.5" />
                ) : (
                  <TrendingUp className="w-5 h-5 text-blue-600 mt-0.5" />
                )
              ) : (
                <AlertTriangle className="w-5 h-5 text-amber-600 mt-0.5" />
              )}
              <div className="text-sm">
                {recommendations.isSafe ? (
                  <>
                    <p className="text-gray-700">
                      {t('weightGoal.recommendation', {
                        action: recommendations.isLosing ? t('weightGoal.lose') : t('weightGoal.gain'),
                        weight: Math.abs(currentWeight - parseFloat(targetWeight)).toFixed(1),
                      })}
                    </p>
                    <p className="text-gray-600 mt-1">
                      • {t('weightGoal.dailyDeficitRecommended', { kcal: recommendations.dailyDeficit })}
                    </p>
                    {recommendations.estimatedDate && (
                      <p className="text-gray-600">
                        • {t('weightGoal.estimatedAt', {
                          date: format(recommendations.estimatedDate, 'PP'),
                        })}
                      </p>
                    )}
                  </>
                ) : (
                  <>
                    <p className="font-medium text-amber-800">{t('weightGoal.warningTitle')}</p>
                    <p className="text-amber-700 mt-1">
                      {t('weightGoal.warningMessage', {
                        rate: recommendations.weeklyRate?.toFixed(1),
                      })}
                    </p>
                  </>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
            {error}
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-2">
          <button
            type="submit"
            disabled={setGoalMutation.isPending || !targetWeight}
            className="px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {setGoalMutation.isPending ? t('common.saving') : (goalProgress ? t('weightGoal.updateGoal') : t('weightGoal.setGoal'))}
          </button>
          {showForm && goalProgress && (
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
            >
              {t('common.cancel')}
            </button>
          )}
        </div>
      </form>
    </div>
  );
}
