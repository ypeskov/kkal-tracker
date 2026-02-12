import {
  CategoryScale,
  Chart as ChartJS,
  ChartOptions,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Title,
  Tooltip,
  TooltipItem,
} from 'chart.js';
import { addDays, differenceInDays, format } from 'date-fns';
import { Line } from 'react-chartjs-2';
import { useTranslation } from 'react-i18next';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
);

interface GoalData {
  targetWeight: number;
  targetDate?: string;
  goalSetAt: string;
  initialWeightAtGoal: number;
  currentWeight: number;
}

interface WeightCaloriesChartProps {
  weightData: { date: string; weight: number }[];
  calorieData: { date: string; calories: number }[];
  showWeight: boolean;
  showCalories: boolean;
  goalData?: GoalData; // Goal data for trajectory line
  showGoalProjection?: boolean; // Extend chart to goal target date with trend projection
}

export default function WeightCaloriesChart({
  weightData,
  calorieData,
  showWeight,
  showCalories,
  goalData,
  showGoalProjection = false,
}: WeightCaloriesChartProps) {
  const { t } = useTranslation();

  // Combine and sort all dates
  const allDates = new Set<string>();
  weightData.forEach(d => allDates.add(d.date));
  calorieData.forEach(d => allDates.add(d.date));
  const sortedDates = Array.from(allDates).sort();

  // Extend dates to goal target date when projection is enabled
  const MAX_PROJECTION_DAYS = 730;
  if (showGoalProjection && goalData?.targetDate && sortedDates.length > 0) {
    const lastDataDate = sortedDates[sortedDates.length - 1];
    const targetDateStr = goalData.targetDate.split('T')[0]; // Normalize to YYYY-MM-DD
    if (targetDateStr > lastDataDate) {
      let current = addDays(new Date(lastDataDate), 1);
      const target = new Date(targetDateStr);
      let daysAdded = 0;
      while (current <= target && daysAdded < MAX_PROJECTION_DAYS) {
        sortedDates.push(format(current, 'yyyy-MM-dd'));
        current = addDays(current, 1);
        daysAdded++;
      }
    }
  }

  // Create maps for quick lookup
  const weightMap = new Map(weightData.map(d => [d.date, d.weight]));
  const calorieMap = new Map(calorieData.map(d => [d.date, d.calories]));

  // Calculate trajectory based on average weight change rate
  const calculateTrajectory = () => {
    if (!goalData || sortedDates.length === 0) return null;

    const goalStartDate = new Date(goalData.goalSetAt);
    const today = new Date();
    const daysSinceGoalStart = differenceInDays(today, goalStartDate);

    if (daysSinceGoalStart <= 0) return null;

    // Calculate average daily change based on actual progress
    const weightChange = goalData.currentWeight - goalData.initialWeightAtGoal;
    const avgDailyChange = weightChange / daysSinceGoalStart;

    // Generate trajectory points for each date in the chart
    return sortedDates.map(dateStr => {
      const date = new Date(dateStr);
      const daysFromGoalStart = differenceInDays(date, goalStartDate);

      if (daysFromGoalStart < 0) return null;

      // Project weight based on average rate
      const projectedWeight = goalData.initialWeightAtGoal + (avgDailyChange * daysFromGoalStart);
      return projectedWeight;
    });
  };

  // Prepare datasets
  const datasets = [];

  if (showWeight) {
    datasets.push({
      label: t('report.weight'),
      data: sortedDates.map(date => weightMap.get(date) || null),
      borderColor: 'rgb(59, 130, 246)',
      backgroundColor: 'rgba(59, 130, 246, 0.5)',
      yAxisID: 'y1',
      tension: 0.1,
      spanGaps: true,
    });

    // Add trajectory line based on average progress
    if (goalData && sortedDates.length > 0) {
      const trajectoryData = calculateTrajectory();
      if (trajectoryData) {
        datasets.push({
          label: t('weightGoal.trajectory'),
          data: trajectoryData,
          borderColor: 'rgb(249, 115, 22)', // Orange
          backgroundColor: 'transparent',
          borderDash: [8, 4],
          borderWidth: 2,
          yAxisID: 'y1',
          pointRadius: 0,
          pointHoverRadius: 0,
          fill: false,
        });
      }

      // Add horizontal target weight line
      datasets.push({
        label: t('weightGoal.targetLine'),
        data: sortedDates.map(() => goalData.targetWeight),
        borderColor: 'rgb(34, 197, 94)',
        backgroundColor: 'transparent',
        borderDash: [3, 3],
        borderWidth: 1,
        yAxisID: 'y1',
        pointRadius: 0,
        pointHoverRadius: 0,
        fill: false,
      });
    }
  }

  if (showCalories) {
    datasets.push({
      label: t('report.calories'),
      data: sortedDates.map(date => {
        const calories = calorieMap.get(date);
        return calories ? calories / 10 : null;
      }),
      borderColor: 'rgb(34, 197, 94)',
      backgroundColor: 'rgba(34, 197, 94, 0.5)',
      yAxisID: 'y2',
      tension: 0.1,
      spanGaps: true,
    });
  }

  const data = {
    labels: sortedDates.map(date => {
      try {
        return format(new Date(date), 'MMM d');
      } catch {
        return date;
      }
    }),
    datasets,
  };

  const options: ChartOptions<'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      mode: 'index' as const,
      intersect: false,
    },
    plugins: {
      legend: {
        position: 'top' as const,
      },
      tooltip: {
        callbacks: {
          label: function (context: TooltipItem<'line'>) {
            let label = context.dataset.label || '';
            if (label) {
              label += ': ';
            }
            if (context.parsed.y !== null) {
              if (context.dataset.yAxisID === 'y1') {
                label += `${context.parsed.y.toFixed(2)} kg`;
              } else {
                // Show actual calories value (multiply back by 10)
                const actualCalories = context.parsed.y * 10;
                label += `${actualCalories.toFixed(0)} kcal`;
              }
            }
            return label;
          },
        },
      },
    },
    scales: {
      x: {
        grid: {
          display: true,
          color: 'rgba(0, 0, 0, 0.05)',
        },
      },
      y1: {
        type: 'linear' as const,
        display: showWeight,
        position: 'left' as const,
        title: {
          display: true,
          text: t('report.weight_kg'),
        },
        grid: {
          color: 'rgba(59, 130, 246, 0.1)',
        },
      },
      y2: {
        type: 'linear' as const,
        display: showCalories,
        position: 'right' as const,
        title: {
          display: true,
          text: `${t('report.calories_kcal')} (÷10)`,
        },
        ticks: {
          callback: function (value: any) {
            return `10×${value.toFixed(1)}`;
          },
        },
        grid: {
          drawOnChartArea: false,
          color: 'rgba(34, 197, 94, 0.1)',
        },
      },
    },
  };

  if (!showWeight && !showCalories) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-400">
        {t('report.select_metric')}
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      <Line data={data} options={options} />
    </div>
  );
}