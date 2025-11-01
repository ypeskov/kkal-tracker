import { useTranslation } from 'react-i18next';

type Period = 'daily' | 'weekly' | 'monthly' | 'yearly';
type StepInterval = '1' | '5' | '10' | '15';

interface ReportFiltersProps {
  dateFrom: string;
  dateTo: string;
  onDateFromChange: (date: string) => void;
  onDateToChange: (date: string) => void;
  activeTab: 'chart' | 'weight';
  period?: Period;
  onPeriodChange?: (period: Period) => void;
  stepInterval?: StepInterval;
  onStepIntervalChange?: (interval: StepInterval) => void;
  showWeight?: boolean;
  onShowWeightChange?: (show: boolean) => void;
  showCalories?: boolean;
  onShowCaloriesChange?: (show: boolean) => void;
  weightStats?: { min: number; max: number; average: number };
}

export default function ReportFilters({
  dateFrom,
  dateTo,
  onDateFromChange,
  onDateToChange,
  activeTab,
  period,
  onPeriodChange,
  stepInterval,
  onStepIntervalChange,
  showWeight,
  onShowWeightChange,
  showCalories,
  onShowCaloriesChange,
  weightStats,
}: ReportFiltersProps) {
  const { t } = useTranslation();

  return (
    <div className="bg-white rounded-lg shadow-md p-4 mb-4">
      {/* All controls in one row on desktop, stacked on mobile */}
      <div className="flex flex-col lg:flex-row lg:items-end gap-3 lg:gap-4">
        {/* From Date */}
        <div className="flex-shrink-0">
          <label className="block text-sm font-medium mb-1">
            {t('report.date_from')}
          </label>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => onDateFromChange(e.target.value)}
            className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 py-3 rounded-lg shadow-sm"
          />
        </div>

        {/* To Date */}
        <div className="flex-shrink-0">
          <label className="block text-sm font-medium mb-1">
            {t('report.date_to')}
          </label>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => onDateToChange(e.target.value)}
            className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 py-3 rounded-lg shadow-sm"
          />
        </div>

        {/* Chart controls */}
        {activeTab === 'chart' && (
          <>
            {/* Period dropdown */}
            {period !== undefined && onPeriodChange && (
              <div className="flex-shrink-0">
                <label className="block text-sm font-medium mb-1">
                  {t('report.period')}
                </label>
                <select
                  value={period}
                  onChange={(e) => onPeriodChange(e.target.value as Period)}
                  className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 pt-4 pb-3 rounded-lg shadow-sm"
                >
                  <option value="daily">{t('report.daily')}</option>
                  <option value="weekly">{t('report.weekly')}</option>
                  <option value="monthly">{t('report.monthly')}</option>
                  <option value="yearly">{t('report.yearly')}</option>
                </select>
              </div>
            )}

            {/* Step Interval dropdown - only visible for daily period */}
            {period === 'daily' && stepInterval !== undefined && onStepIntervalChange && (
              <div className="flex-shrink-0">
                <label className="block text-sm font-medium mb-1">
                  {t('report.step_interval')}
                </label>
                <select
                  value={stepInterval}
                  onChange={(e) => onStepIntervalChange(e.target.value as StepInterval)}
                  className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 pt-4 pb-3 rounded-lg shadow-sm"
                >
                  <option value="1">{t('report.step_1_day')}</option>
                  <option value="5">{t('report.step_5_days')}</option>
                  <option value="10">{t('report.step_10_days')}</option>
                  <option value="15">{t('report.step_15_days')}</option>
                </select>
              </div>
            )}

            {/* Checkboxes */}
            {showWeight !== undefined && showCalories !== undefined &&
             onShowWeightChange && onShowCaloriesChange && (
              <div className="flex flex-col lg:flex-row gap-3 lg:gap-4 lg:items-center lg:h-[58px]">
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={showWeight}
                    onChange={(e) => onShowWeightChange(e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm font-medium">{t('report.weight')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={showCalories}
                    onChange={(e) => onShowCaloriesChange(e.target.checked)}
                    className="w-4 h-4 text-green-600 rounded"
                  />
                  <span className="text-sm font-medium">{t('report.calories')}</span>
                </div>
              </div>
            )}
          </>
        )}

        {/* Weight stats for weight history tab */}
        {activeTab === 'weight' && weightStats && weightStats.min > 0 && (
          <div className="flex flex-col lg:flex-row gap-3 lg:gap-4 lg:ml-auto w-full lg:w-auto">
            <div className="flex flex-row gap-3 sm:gap-4 justify-center lg:justify-start">
              <div className="flex flex-col">
                <span className="text-xs font-medium text-gray-600">{t('report.min_weight')}</span>
                <span className="text-lg font-bold text-blue-600">{weightStats.min} kg</span>
              </div>
              <div className="flex flex-col">
                <span className="text-xs font-medium text-gray-600">{t('report.max_weight')}</span>
                <span className="text-lg font-bold text-blue-600">{weightStats.max} kg</span>
              </div>
              <div className="flex flex-col">
                <span className="text-xs font-medium text-gray-600">{t('report.avg_weight')}</span>
                <span className="text-lg font-bold text-blue-600">{weightStats.average} kg</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
