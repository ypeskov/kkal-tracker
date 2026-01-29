import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { format, startOfWeek, startOfMonth, startOfYear } from 'date-fns';
import WeightCaloriesChart from '@/components/reports/WeightCaloriesChart';
import WeightHistory from '@/components/reports/WeightHistory';
import TabNavigation from '@/components/TabNavigation';
import ReportFilters from '@/components/ReportFilters';
import StatisticsCard from '@/components/StatisticsCard';
import { reportsService } from '@/api/reports';
import { profileService } from '@/api/profile';
import { BarChart3, Weight } from 'lucide-react';

type Period = 'daily' | 'weekly' | 'monthly' | 'yearly';
type StepInterval = '1' | '5' | '10' | '15';

export default function Report() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'chart' | 'weight'>('chart');
  const [showWeight, setShowWeight] = useState(true);
  const [showCalories, setShowCalories] = useState(true);
  const [period, setPeriod] = useState<Period>('daily');
  const [stepInterval, setStepInterval] = useState<StepInterval>('1');

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

  // Fetch weight goal progress
  const { data: goalProgress } = useQuery({
    queryKey: ['weightGoalProgress'],
    queryFn: profileService.getWeightGoalProgress,
  });

  // Aggregate data based on selected period
  const aggregatedData = useMemo(() => {
    if (!reportData) return { weightData: [], calorieData: [] };

    const { weight_history, calorie_history } = reportData;

    if (period === 'daily') {
      // Apply step interval for daily period
      const step = parseInt(stepInterval);
      if (step === 1) {
        return {
          weightData: weight_history,
          calorieData: calorie_history,
        };
      }

      // Filter data by step interval
      const filteredWeightData: typeof weight_history = [];
      const filteredCalorieData: typeof calorie_history = [];

      // Sort data by date first
      const sortedWeightHistory = [...weight_history].sort((a, b) => a.date.localeCompare(b.date));
      const sortedCalorieHistory = [...calorie_history].sort((a, b) => a.date.localeCompare(b.date));

      // Get all unique dates from both datasets
      const allDates = new Set<string>();
      sortedWeightHistory.forEach(d => allDates.add(d.date));
      sortedCalorieHistory.forEach(d => allDates.add(d.date));
      const sortedDates = Array.from(allDates).sort();

      // Pick dates at step intervals
      for (let i = 0; i < sortedDates.length; i += step) {
        const date = sortedDates[i];
        const weightItem = sortedWeightHistory.find(w => w.date === date);
        const calorieItem = sortedCalorieHistory.find(c => c.date === date);

        if (weightItem) filteredWeightData.push(weightItem);
        if (calorieItem) filteredCalorieData.push(calorieItem);
      }

      // Always include the last date if it wasn't already included
      const lastDate = sortedDates[sortedDates.length - 1];
      const lastIndex = sortedDates.length - 1;
      if (lastIndex % step !== 0 && lastDate) {
        const lastWeightItem = sortedWeightHistory.find(w => w.date === lastDate);
        const lastCalorieItem = sortedCalorieHistory.find(c => c.date === lastDate);
        if (lastWeightItem && !filteredWeightData.some(w => w.date === lastDate)) {
          filteredWeightData.push(lastWeightItem);
        }
        if (lastCalorieItem && !filteredCalorieData.some(c => c.date === lastDate)) {
          filteredCalorieData.push(lastCalorieItem);
        }
      }

      return {
        weightData: filteredWeightData,
        calorieData: filteredCalorieData,
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
  }, [reportData, period, stepInterval]);

  // Calculate weight statistics from raw data (single source of truth for both tabs)
  const weightStats = useMemo(() => {
    if (!reportData?.weight_history || reportData.weight_history.length === 0) {
      return { min: 0, max: 0, average: 0 };
    }

    const weights = reportData.weight_history.map(item => item.weight);
    const min = Math.min(...weights);
    const max = Math.max(...weights);
    const average = weights.reduce((sum, w) => sum + w, 0) / weights.length;

    return {
      min: Number(min.toFixed(2)),
      max: Number(max.toFixed(2)),
      average: Number(average.toFixed(2)),
    };
  }, [reportData?.weight_history]);

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
    <div className="px-4 py-2 md:px-6 lg:px-8">
      <div className="mb-6">
        <h2 className="text-3xl font-semibold text-gray-800">{t('report.title')}</h2>
      </div>

      {/* Tabs */}
      <TabNavigation
        tabs={[
          { id: 'chart', label: t('report.charts_tab'), icon: <BarChart3 size={18} /> },
          { id: 'weight', label: t('report.weight_history_tab'), icon: <Weight size={18} /> },
        ]}
        activeTab={activeTab}
        onTabChange={(tabId) => setActiveTab(tabId as 'chart' | 'weight')}
      />

      {/* Date Range Filters */}
      <ReportFilters
        dateFrom={dateFrom}
        dateTo={dateTo}
        onDateFromChange={setDateFrom}
        onDateToChange={setDateTo}
        activeTab={activeTab}
        period={period}
        onPeriodChange={setPeriod}
        stepInterval={stepInterval}
        onStepIntervalChange={setStepInterval}
        showWeight={showWeight}
        onShowWeightChange={setShowWeight}
        showCalories={showCalories}
        onShowCaloriesChange={setShowCalories}
        weightStats={weightStats}
      />

      {/* Tab Content */}
      {activeTab === 'chart' ? (
        <div className="space-y-4">
          {/* Weight and Calories Statistics */}
          {(showWeight && reportData?.weight_history && reportData.weight_history.length > 0) || avgCaloriesPerDay > 0 ? (
            <StatisticsCard
              stats={[
                ...(showWeight && reportData?.weight_history && reportData.weight_history.length > 0
                  ? [
                      { label: t('report.min_weight'), value: `${weightStats.min} kg`, color: 'text-blue-600' },
                      { label: t('report.max_weight'), value: `${weightStats.max} kg`, color: 'text-blue-600' },
                      { label: t('report.avg_weight'), value: `${weightStats.average} kg`, color: 'text-blue-600' },
                    ]
                  : []),
                ...(avgCaloriesPerDay > 0
                  ? [{ label: t('report.avg_calories_per_day'), value: `${avgCaloriesPerDay} kcal`, color: 'text-green-600' }]
                  : []),
              ]}
            />
          ) : null}

          {/* Chart */}
          <div className="bg-white rounded-lg shadow-md p-4">
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
                  goalData={goalProgress ? {
                    targetWeight: goalProgress.target_weight,
                    targetDate: goalProgress.target_date,
                    goalSetAt: goalProgress.goal_set_at,
                    initialWeightAtGoal: goalProgress.initial_weight_at_goal,
                    currentWeight: goalProgress.current_weight,
                  } : undefined}
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