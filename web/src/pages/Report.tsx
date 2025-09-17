import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { format, startOfWeek, startOfMonth, startOfYear } from 'date-fns';
import WeightCaloriesChart from '../components/reports/WeightCaloriesChart';
import WeightHistory from '../components/reports/WeightHistory';
import { reportsService } from '../api/reports';
import { BarChart3, Weight } from 'lucide-react';

type Period = 'daily' | 'weekly' | 'monthly' | 'yearly';

export default function Report() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'chart' | 'weight'>('chart');
  const [showWeight, setShowWeight] = useState(true);
  const [showCalories, setShowCalories] = useState(true);
  const [period, setPeriod] = useState<Period>('daily');

  // Default date range: last 30 days
  const [dateFrom, setDateFrom] = useState(
    format(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), 'yyyy-MM-dd')
  );
  const [dateTo, setDateTo] = useState(
    format(new Date(), 'yyyy-MM-dd')
  );

  // Fetch report data
  const { data: reportData, isLoading } = useQuery({
    queryKey: ['reportData', dateFrom, dateTo],
    queryFn: () => reportsService.getReportData(dateFrom, dateTo),
  });

  // Aggregate data based on selected period
  const aggregatedData = useMemo(() => {
    if (!reportData) return { weightData: [], calorieData: [] };

    const { weight_history, calorie_history } = reportData;

    if (period === 'daily') {
      return {
        weightData: weight_history,
        calorieData: calorie_history,
      };
    }

    // Group data by period
    const weightGroups: { [key: string]: number[] } = {};
    const calorieGroups: { [key: string]: number[] } = {};

    // Process weight data
    weight_history.forEach(item => {
      const date = new Date(item.date);
      let key: string;

      switch (period) {
        case 'weekly':
          const weekStart = startOfWeek(date, { weekStartsOn: 1 }); // Monday start
          key = format(weekStart, 'yyyy-MM-dd');
          break;
        case 'monthly':
          key = format(startOfMonth(date), 'yyyy-MM-dd');
          break;
        case 'yearly':
          key = format(startOfYear(date), 'yyyy-MM-dd');
          break;
        default:
          key = item.date;
      }

      if (!weightGroups[key]) weightGroups[key] = [];
      weightGroups[key].push(item.weight);
    });

    // Process calorie data
    calorie_history.forEach(item => {
      const date = new Date(item.date);
      let key: string;

      switch (period) {
        case 'weekly':
          const weekStart = startOfWeek(date, { weekStartsOn: 1 }); // Monday start
          key = format(weekStart, 'yyyy-MM-dd');
          break;
        case 'monthly':
          key = format(startOfMonth(date), 'yyyy-MM-dd');
          break;
        case 'yearly':
          key = format(startOfYear(date), 'yyyy-MM-dd');
          break;
        default:
          key = item.date;
      }

      if (!calorieGroups[key]) calorieGroups[key] = [];
      calorieGroups[key].push(item.calories);
    });

    // Calculate averages for each period
    const weightData = Object.entries(weightGroups).map(([date, weights]) => ({
      date,
      weight: weights.reduce((sum, w) => sum + w, 0) / weights.length,
    }));

    const calorieData = Object.entries(calorieGroups).map(([date, calories]) => ({
      date,
      calories: Math.round(calories.reduce((sum, c) => sum + c, 0) / calories.length),
    }));

    return {
      weightData: weightData.sort((a, b) => a.date.localeCompare(b.date)),
      calorieData: calorieData.sort((a, b) => a.date.localeCompare(b.date)),
    };
  }, [reportData, period]);

  // Calculate weight statistics for the aggregated data
  const weightStats = useMemo(() => {
    if (!aggregatedData.weightData || aggregatedData.weightData.length === 0) {
      return { min: 0, max: 0, average: 0 };
    }

    const weights = aggregatedData.weightData.map(item => item.weight);
    const min = Math.min(...weights);
    const max = Math.max(...weights);
    const average = weights.reduce((sum, w) => sum + w, 0) / weights.length;

    return {
      min: Number(min.toFixed(2)),
      max: Number(max.toFixed(2)),
      average: Number(average.toFixed(2)),
    };
  }, [aggregatedData.weightData]);

  // Calculate average calories per day
  const avgCaloriesPerDay = useMemo(() => {
    if (!reportData || !reportData.calorie_history || reportData.calorie_history.length === 0) {
      return 0;
    }

    // Get all unique dates from calorie history
    const uniqueDates = new Set(reportData.calorie_history.map(item => item.date));
    const totalDays = uniqueDates.size;

    // Calculate total calories
    const totalCalories = reportData.calorie_history.reduce((sum, item) => sum + item.calories, 0);

    // Return average calories per day
    return Math.round(totalCalories / totalDays);
  }, [reportData]);

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">{t('report.title')}</h2>
      </div>

      {/* Tabs */}
      <div className="flex flex-col sm:flex-row gap-2 mb-4">
        <button
          onClick={() => setActiveTab('chart')}
          className={`flex items-center justify-center gap-2 px-4 py-2 rounded-lg transition-colors ${
            activeTab === 'chart'
              ? 'bg-blue-600 text-white'
              : 'bg-white text-gray-700 hover:bg-gray-100'
          }`}
        >
          <BarChart3 size={18} />
          <span>{t('report.charts_tab')}</span>
        </button>
        <button
          onClick={() => setActiveTab('weight')}
          className={`flex items-center justify-center gap-2 px-4 py-2 rounded-lg transition-colors ${
            activeTab === 'weight'
              ? 'bg-blue-600 text-white'
              : 'bg-white text-gray-700 hover:bg-gray-100'
          }`}
        >
          <Weight size={18} />
          <span>{t('report.weight_history_tab')}</span>
        </button>
      </div>

      {/* Date Range Filters */}
      <div className="card p-4 mb-4">
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
              onChange={(e) => setDateFrom(e.target.value)}
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
              onChange={(e) => setDateTo(e.target.value)}
              className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 py-3 rounded-lg shadow-sm"
            />
          </div>

          {/* Chart controls */}
          {activeTab === 'chart' && (
            <>
              {/* Period dropdown */}
              <div className="flex-shrink-0">
                <label className="block text-sm font-medium mb-1">
                  {t('report.period')}
                </label>
                <select
                  value={period}
                  onChange={(e) => setPeriod(e.target.value as Period)}
                  className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 py-3 rounded-lg shadow-sm"
                >
                  <option value="daily">{t('report.daily')}</option>
                  <option value="weekly">{t('report.weekly')}</option>
                  <option value="monthly">{t('report.monthly')}</option>
                  <option value="yearly">{t('report.yearly')}</option>
                </select>
              </div>

              {/* Checkboxes */}
              <div className="flex flex-col lg:flex-row gap-3 lg:gap-4 lg:items-center lg:h-[58px]">
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={showWeight}
                    onChange={(e) => setShowWeight(e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm font-medium">{t('report.weight')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={showCalories}
                    onChange={(e) => setShowCalories(e.target.checked)}
                    className="w-4 h-4 text-green-600 rounded"
                  />
                  <span className="text-sm font-medium">{t('report.calories')}</span>
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'chart' ? (
        <div className="space-y-4">
          {/* Weight and Calories Statistics */}
          {(showWeight && aggregatedData.weightData.length > 0) || avgCaloriesPerDay > 0 ? (
            <div className="card p-4">
              <div className="flex flex-col sm:flex-row gap-4 sm:gap-8 justify-center text-center">
                {showWeight && aggregatedData.weightData.length > 0 && (
                  <>
                    <div className="flex flex-col">
                      <span className="text-sm font-medium text-gray-600">{t('report.min_weight')}</span>
                      <span className="text-xl font-bold text-blue-600">{weightStats.min} kg</span>
                    </div>
                    <div className="flex flex-col">
                      <span className="text-sm font-medium text-gray-600">{t('report.max_weight')}</span>
                      <span className="text-xl font-bold text-blue-600">{weightStats.max} kg</span>
                    </div>
                    <div className="flex flex-col">
                      <span className="text-sm font-medium text-gray-600">{t('report.avg_weight')}</span>
                      <span className="text-xl font-bold text-blue-600">{weightStats.average} kg</span>
                    </div>
                  </>
                )}
                {avgCaloriesPerDay > 0 && (
                  <div className="flex flex-col">
                    <span className="text-sm font-medium text-gray-600">{t('report.avg_calories_per_day')}</span>
                    <span className="text-xl font-bold text-green-600">{avgCaloriesPerDay} kcal</span>
                  </div>
                )}
              </div>
            </div>
          ) : null}

          {/* Chart */}
          <div className="card p-4">
            <div className="h-64 sm:h-96">
              {isLoading ? (
                <div className="flex items-center justify-center h-full">
                  <span className="text-gray-500">{t('common.loading')}</span>
                </div>
              ) : (
                <WeightCaloriesChart
                  weightData={aggregatedData.weightData}
                  calorieData={aggregatedData.calorieData}
                  showWeight={showWeight}
                  showCalories={showCalories}
                />
              )}
            </div>
          </div>
        </div>
      ) : (
        <WeightHistory dateFrom={dateFrom} dateTo={dateTo} />
      )}
    </div>
  );
}