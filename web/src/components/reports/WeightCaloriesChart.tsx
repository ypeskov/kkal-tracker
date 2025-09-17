import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  ChartOptions,
  TooltipItem,
} from 'chart.js';
import { useTranslation } from 'react-i18next';
import { format } from 'date-fns';

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

interface WeightCaloriesChartProps {
  weightData: { date: string; weight: number }[];
  calorieData: { date: string; calories: number }[];
  showWeight: boolean;
  showCalories: boolean;
}

export default function WeightCaloriesChart({
  weightData,
  calorieData,
  showWeight,
  showCalories,
}: WeightCaloriesChartProps) {
  const { t } = useTranslation();

  // Combine and sort all dates
  const allDates = new Set<string>();
  weightData.forEach(d => allDates.add(d.date));
  calorieData.forEach(d => allDates.add(d.date));
  const sortedDates = Array.from(allDates).sort();

  // Create maps for quick lookup
  const weightMap = new Map(weightData.map(d => [d.date, d.weight]));
  const calorieMap = new Map(calorieData.map(d => [d.date, d.calories]));

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
          label: function(context: TooltipItem<'line'>) {
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
          callback: function(value: any) {
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