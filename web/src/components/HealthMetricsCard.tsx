import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { metricsService } from '@/api/metrics';
import { Activity } from 'lucide-react';

export default function HealthMetricsCard() {
  const { t } = useTranslation();

  const { data: healthMetrics, isLoading } = useQuery({
    queryKey: ['healthMetrics'],
    queryFn: metricsService.getHealthMetrics,
  });

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-4">
        <div className="flex items-center gap-2 mb-4">
          <Activity className="w-5 h-5 text-blue-600" />
          <h3 className="text-lg font-semibold text-gray-800">{t('profile.healthMetrics')}</h3>
        </div>
        <div className="text-gray-500 text-sm">{t('common.loading')}</div>
      </div>
    );
  }

  if (!healthMetrics || (!healthMetrics.bmi && !healthMetrics.bmr)) {
    return null;
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-4">
      <div className="flex items-center gap-2 mb-4">
        <Activity className="w-5 h-5 text-blue-600" />
        <h3 className="text-lg font-semibold text-gray-800">{t('profile.healthMetrics')}</h3>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        {healthMetrics.bmi && (
          <div className="p-3 bg-blue-50 rounded-lg">
            <div className="text-xs text-gray-600 mb-1">{t('profile.bmi')}</div>
            <div className="text-xl font-bold text-blue-600">{healthMetrics.bmi.toFixed(1)}</div>
            {healthMetrics.bmi_category && (
              <div className="text-xs text-gray-500 mt-1">
                {t(`profile.bmiCategory.${healthMetrics.bmi_category}`)}
              </div>
            )}
          </div>
        )}
        
        {healthMetrics.bmr && (
          <div className="p-3 bg-green-50 rounded-lg">
            <div className="text-xs text-gray-600 mb-1">{t('profile.bmr')}</div>
            <div className="text-xl font-bold text-green-600">{Math.round(healthMetrics.bmr)}</div>
            <div className="text-xs text-gray-500 mt-1">{t('common.kcal')}/day</div>
          </div>
        )}

        {healthMetrics.tdee && (
          <div className="p-3 bg-purple-50 rounded-lg">
            <div className="text-xs text-gray-600 mb-1">{t('profile.tdee')}</div>
            <div className="text-xl font-bold text-purple-600">{Math.round(healthMetrics.tdee)}</div>
            <div className="text-xs text-gray-500 mt-1">{t('common.kcal')}/day</div>
          </div>
        )}
      </div>

      {healthMetrics.health_status && (
        <div className="mt-3 p-2 bg-gray-50 rounded text-xs text-gray-600">
          {t(`profile.healthStatus.${healthMetrics.health_status}`)}
        </div>
      )}
    </div>
  );
}
