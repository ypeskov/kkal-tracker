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
    <div className="card mb-lg p-lg">
      <div className="mb-lg">
        <label htmlFor="filterType" className="form-group label font-bold mb-sm">
          {t('dashboard.filterBy')}:
        </label>
        <select
          id="filterType"
          value={filterType}
          onChange={(e) => onFilterChange(e.target.value as FilterType)}
          className="form-input mr-lg"
        >
          <option value="today">{t('dashboard.today')}</option>
          <option value="yesterday">{t('dashboard.yesterday')}</option>
          <option value="lastWeek">{t('dashboard.lastWeek')}</option>
          <option value="lastMonth">{t('dashboard.lastMonth')}</option>
          <option value="customRange">{t('dashboard.customRange')}</option>
        </select>
      </div>
      
      {filterType === 'customRange' && (
        <div className="flex gap-lg items-center flex-wrap">
          <div>
            <label htmlFor="dateFrom" className="form-group label mb-xs">
              {t('dashboard.dateFrom')}:
            </label>
            <input
              type="date"
              id="dateFrom"
              value={customDateFrom}
              onChange={(e) => onCustomDateFromChange(e.target.value)}
              className="form-input"
            />
          </div>
          <div>
            <label htmlFor="dateTo" className="form-group label mb-xs">
              {t('dashboard.dateTo')}:
            </label>
            <input
              type="date"
              id="dateTo"
              value={customDateTo}
              onChange={(e) => onCustomDateToChange(e.target.value)}
              className="form-input"
            />
          </div>
        </div>
      )}
      
      <div className="mt-lg p-md bg-success-light rounded-sm font-bold text-success-dark">
        <div className="mb-sm">
          {t('dashboard.totalCaloriesFiltered')} {nutritionTotals.calories} {t('dashboard.kcal')}
        </div>
        {(nutritionTotals.fats > 0 || nutritionTotals.carbs > 0 || nutritionTotals.proteins > 0) && (
          <div className="flex gap-lg text-sm flex-wrap">
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