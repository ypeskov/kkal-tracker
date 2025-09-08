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
    <div className="filter-section" style={{ marginBottom: '1rem', padding: '1rem', backgroundColor: '#f8f9fa', borderRadius: '8px' }}>
      <div style={{ marginBottom: '1rem' }}>
        <label htmlFor="filterType" style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
          {t('dashboard.filterBy')}:
        </label>
        <select 
          id="filterType"
          value={filterType} 
          onChange={(e) => onFilterChange(e.target.value as FilterType)}
          style={{ padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd', marginRight: '1rem' }}
        >
          <option value="today">{t('dashboard.today')}</option>
          <option value="yesterday">{t('dashboard.yesterday')}</option>
          <option value="lastWeek">{t('dashboard.lastWeek')}</option>
          <option value="lastMonth">{t('dashboard.lastMonth')}</option>
          <option value="customRange">{t('dashboard.customRange')}</option>
        </select>
      </div>
      
      {filterType === 'customRange' && (
        <div className="custom-date-range" style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <div>
            <label htmlFor="dateFrom" style={{ display: 'block', marginBottom: '0.25rem' }}>
              {t('dashboard.dateFrom')}:
            </label>
            <input
              type="date"
              id="dateFrom"
              value={customDateFrom}
              onChange={(e) => onCustomDateFromChange(e.target.value)}
              style={{ padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd' }}
            />
          </div>
          <div>
            <label htmlFor="dateTo" style={{ display: 'block', marginBottom: '0.25rem' }}>
              {t('dashboard.dateTo')}:
            </label>
            <input
              type="date"
              id="dateTo"
              value={customDateTo}
              onChange={(e) => onCustomDateToChange(e.target.value)}
              style={{ padding: '0.5rem', borderRadius: '4px', border: '1px solid #ddd' }}
            />
          </div>
        </div>
      )}
      
      <div style={{ marginTop: '1rem', padding: '0.75rem', backgroundColor: '#e8f5e8', borderRadius: '4px', fontWeight: 'bold', color: '#2d5016' }}>
        <div style={{ marginBottom: '0.5rem' }}>
          {t('dashboard.totalCaloriesFiltered')} {nutritionTotals.calories} {t('dashboard.kcal')}
        </div>
        {(nutritionTotals.fats > 0 || nutritionTotals.carbs > 0 || nutritionTotals.proteins > 0) && (
          <div className="nutrition-summary" style={{ display: 'flex', gap: '1rem', fontSize: '0.9em', flexWrap: 'wrap' }}>
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