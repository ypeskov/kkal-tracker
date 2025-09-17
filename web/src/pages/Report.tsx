import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { format } from 'date-fns';
import WeightCaloriesChart from '../components/reports/WeightCaloriesChart';
import WeightHistory from '../components/reports/WeightHistory';
import { reportsService } from '../api/reports';
import { BarChart3, Weight } from 'lucide-react';

export default function Report() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'chart' | 'weight'>('chart');
  const [showWeight, setShowWeight] = useState(true);
  const [showCalories, setShowCalories] = useState(true);

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
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          <div>
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
          <div>
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

          {/* Toggles for chart tab */}
          {activeTab === 'chart' && (
            <>
              <div className="flex items-end">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={showWeight}
                    onChange={(e) => setShowWeight(e.target.checked)}
                    className="w-4 h-4 text-blue-600 rounded"
                  />
                  <span className="text-sm font-medium">{t('report.show_weight')}</span>
                </label>
              </div>
              <div className="flex items-end">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={showCalories}
                    onChange={(e) => setShowCalories(e.target.checked)}
                    className="w-4 h-4 text-green-600 rounded"
                  />
                  <span className="text-sm font-medium">{t('report.show_calories')}</span>
                </label>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'chart' ? (
        <div className="card p-4">
          <div className="h-64 sm:h-96">
            {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <span className="text-gray-500">{t('common.loading')}</span>
              </div>
            ) : (
              <WeightCaloriesChart
                weightData={reportData?.weight_history || []}
                calorieData={reportData?.calorie_history || []}
                showWeight={showWeight}
                showCalories={showCalories}
              />
            )}
          </div>
        </div>
      ) : (
        <WeightHistory dateFrom={dateFrom} dateTo={dateTo} />
      )}
    </div>
  );
}