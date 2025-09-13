import { useTranslation } from 'react-i18next';

type FilterType = 'today' | 'yesterday' | 'lastWeek' | 'lastMonth' | 'customRange';

interface FilterSectionProps {
  filterType: FilterType;
  onFilterChange: (filter: FilterType) => void;
  customDateFrom: string;
  onCustomDateFromChange: (date: string) => void;
  customDateTo: string;
  onCustomDateToChange: (date: string) => void;
  nutritionTotals: {
    calories: number;
    fats: number;
    carbs: number;
    proteins: number;
  };
}

export default function FilterSection({ 
  filterType, 
  onFilterChange, 
  customDateFrom, 
  onCustomDateFromChange, 
  customDateTo, 
  onCustomDateToChange, 
  nutritionTotals 
}: FilterSectionProps) {
  const { t } = useTranslation();

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6 p-6">
      <div className="mb-6">
        <label htmlFor="filterType" className="block font-bold mb-2 text-gray-800 text-sm">
          {t('dashboard.filterBy')}:
        </label>
        <select
          id="filterType"
          value={filterType}
          onChange={(e) => onFilterChange(e.target.value as FilterType)}
          className="w-full px-3 py-2 mr-6 border border-gray-300 rounded focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-opacity-10 text-sm transition-colors bg-white"
        >
          <option value="today">{t('dashboard.today')}</option>
          <option value="yesterday">{t('dashboard.yesterday')}</option>
          <option value="lastWeek">{t('dashboard.lastWeek')}</option>
          <option value="lastMonth">{t('dashboard.lastMonth')}</option>
          <option value="customRange">{t('dashboard.customRange')}</option>
        </select>
      </div>
      
      {filterType === 'customRange' && (
        <div className="flex gap-6 items-center flex-wrap">
          <div>
            <label htmlFor="dateFrom" className="block mb-1 text-gray-800 text-sm font-medium">
              {t('dashboard.dateFrom')}:
            </label>
            <input
              type="date"
              id="dateFrom"
              value={customDateFrom}
              onChange={(e) => onCustomDateFromChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-opacity-10 text-sm transition-colors bg-white"
            />
          </div>
          <div>
            <label htmlFor="dateTo" className="block mb-1 text-gray-800 text-sm font-medium">
              {t('dashboard.dateTo')}:
            </label>
            <input
              type="date"
              id="dateTo"
              value={customDateTo}
              onChange={(e) => onCustomDateToChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-opacity-10 text-sm transition-colors bg-white"
            />
          </div>
        </div>
      )}
      
      <div className="mt-6 p-4 bg-green-50 rounded border border-green-200 font-bold text-green-800">
        <div className="mb-2">
          {t('dashboard.totalCaloriesFiltered')} {nutritionTotals.calories} {t('dashboard.kcal')}
        </div>
        {(nutritionTotals.fats > 0 || nutritionTotals.carbs > 0 || nutritionTotals.proteins > 0) && (
          <div className="flex gap-6 text-sm flex-wrap">
            {nutritionTotals.fats > 0 && (
              <span>{t('dashboard.fats')}: {nutritionTotals.fats.toFixed(1)}g</span>
            )}
            {nutritionTotals.carbs > 0 && (
              <span>{t('dashboard.carbs')}: {nutritionTotals.carbs.toFixed(1)}g</span>
            )}
            {nutritionTotals.proteins > 0 && (
              <span>{t('dashboard.proteins')}: {nutritionTotals.proteins.toFixed(1)}g</span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}