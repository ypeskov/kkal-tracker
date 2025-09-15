import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import EntryListItem from './EntryListItem';

interface CalorieEntriesListProps {
  entries: any[];
  onEdit: (entry: any) => void;
}

export default function CalorieEntriesList({ entries, onEdit }: CalorieEntriesListProps) {
  const { t, i18n } = useTranslation();
  const [expandedDates, setExpandedDates] = useState<Set<string>>(new Set());

  const entriesByDate = useMemo(() => {
    if (!entries || entries.length === 0) {
      return {};
    }

    const grouped: { [date: string]: any[] } = {};

    entries.forEach(entry => {
      const entryDate = new Date(entry.meal_datetime);
      const year = entryDate.getFullYear();
      const month = String(entryDate.getMonth() + 1).padStart(2, '0');
      const day = String(entryDate.getDate()).padStart(2, '0');
      const date = `${year}-${month}-${day}`;
      if (!grouped[date]) {
        grouped[date] = [];
      }
      grouped[date].push(entry);
    });

    Object.keys(grouped).forEach(date => {
      grouped[date].sort((a, b) => new Date(a.meal_datetime).getTime() - new Date(b.meal_datetime).getTime());
    });

    return grouped;
  }, [entries]);

  const formatDateHeader = (dateString: string) => {
    const date = new Date(dateString);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    // Format dates using local date components
    const formatLocalDate = (d: Date) => {
      const year = d.getFullYear();
      const month = String(d.getMonth() + 1).padStart(2, '0');
      const day = String(d.getDate()).padStart(2, '0');
      return `${year}-${month}-${day}`;
    };

    const dateOnly = formatLocalDate(date);
    const todayOnly = formatLocalDate(today);
    const yesterdayOnly = formatLocalDate(yesterday);

    const currentLang = i18n.language;
    const locale = currentLang === 'uk-UA' ? 'uk-UA' : currentLang === 'ru-UA' ? 'ru-RU' : 'en-US';

    if (dateOnly === todayOnly) {
      return t('dashboard.today') + ' (' + date.toLocaleDateString(locale) + ')';
    } else if (dateOnly === yesterdayOnly) {
      return t('dashboard.yesterday') + ' (' + date.toLocaleDateString(locale) + ')';
    } else {
      return date.toLocaleDateString(locale);
    }
  };

  const calculateDailyTotals = (dateEntries: any[]) => {
    return dateEntries.reduce((totals, entry) => ({
      calories: totals.calories + (entry.calories || 0),
      fats: totals.fats + ((entry.fats || 0) * entry.weight / 100),
      carbs: totals.carbs + ((entry.carbs || 0) * entry.weight / 100),
      proteins: totals.proteins + ((entry.proteins || 0) * entry.weight / 100)
    }), { calories: 0, fats: 0, carbs: 0, proteins: 0 });
  };

  // Check if we have multiple dates (for collapsible behavior)
  const hasMultipleDates = Object.keys(entriesByDate).length > 1;

  const toggleDateExpansion = (dateKey: string) => {
    setExpandedDates(prev => {
      const newSet = new Set(prev);
      if (newSet.has(dateKey)) {
        newSet.delete(dateKey);
      } else {
        newSet.add(dateKey);
      }
      return newSet;
    });
  };

  return (
    <section className="mb-8">
      <h2 className="text-xl font-semibold text-gray-800 mb-4">{t('dashboard.todayEntries')}</h2>
      {entries?.length === 0 ? (
        <p className="text-center py-8 text-gray-600">{t('dashboard.noEntries')}</p>
      ) : (
        <div className="flex flex-col gap-6">
          {Object.keys(entriesByDate)
            .sort((a, b) => new Date(b).getTime() - new Date(a).getTime()) // Sort dates desc (newest first)
            .map((dateKey, index) => {
              const isExpanded = !hasMultipleDates || expandedDates.has(dateKey);
              const dailyTotals = calculateDailyTotals(entriesByDate[dateKey]);
              
              return (
                <div key={dateKey} className={`rounded-lg p-4 mb-6 border shadow-md relative transition-all duration-200 hover:translate-x-0.5 ${
                  index % 2 === 0
                    ? 'bg-blue-50 border-blue-200'
                    : 'bg-green-50 border-green-200'
                }`}>
                  {/* Clickable date header with total */}
                  <div
                    onClick={() => hasMultipleDates && toggleDateExpansion(dateKey)}
                    className={`m-0 p-4 text-lg font-semibold rounded-md transition-all duration-200 ${
                      index % 2 === 0
                        ? 'text-blue-800 bg-gradient-to-r from-white to-blue-50 border-l-4 border-blue-500 shadow-blue-200'
                        : 'text-green-800 bg-gradient-to-r from-white to-green-50 border-l-4 border-green-500 shadow-green-200'
                    } ${
                      hasMultipleDates ? 'cursor-pointer select-none hover:translate-x-1 hover:shadow-lg active:translate-x-0.5' : ''
                    } ${
                      isExpanded ? 'mb-4' : 'mb-0'
                    } shadow-md`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {hasMultipleDates && (
                          <span className={`inline-block w-5 h-5 leading-5 text-center text-sm font-bold transition-transform duration-200 ${
                            isExpanded ? 'rotate-0' : '-rotate-90'
                          }`}>
                            ▼
                          </span>
                        )}
                        <span>{formatDateHeader(dateKey)}</span>
                      </div>
                      <span className="text-green-600 font-bold text-base">
                        {dailyTotals.calories} {t('dashboard.kcal')}
                      </span>
                    </div>
                    
                    {/* Show nutrient summary in header when collapsed */}
                    {!isExpanded && (dailyTotals.fats > 0 || dailyTotals.carbs > 0 || dailyTotals.proteins > 0) && (
                      <div className="flex md:flex-row md:gap-6 sm:flex-wrap sm:gap-2 text-sm mt-2 opacity-80">
                        {dailyTotals.fats > 0 && (
                          <span className="md:flex-1 md:bg-transparent md:p-0 md:rounded-none md:text-inherit md:leading-normal sm:flex-[0_0_calc(50%-0.25rem)] sm:text-center sm:p-1 sm:bg-white/30 sm:rounded sm:text-xs sm:leading-tight">{t('dashboard.fatsConsumed')}: {dailyTotals.fats.toFixed(1)}g</span>
                        )}
                        {dailyTotals.carbs > 0 && (
                          <span className="md:flex-1 md:bg-transparent md:p-0 md:rounded-none md:text-inherit md:leading-normal sm:flex-[0_0_calc(50%-0.25rem)] sm:text-center sm:p-1 sm:bg-white/30 sm:rounded sm:text-xs sm:leading-tight">{t('dashboard.carbsConsumed')}: {dailyTotals.carbs.toFixed(1)}g</span>
                        )}
                        {dailyTotals.proteins > 0 && (
                          <span className="md:flex-1 md:bg-transparent md:p-0 md:rounded-none md:text-inherit md:leading-normal sm:flex-[0_0_calc(50%-0.25rem)] sm:text-center sm:p-1 sm:bg-white/30 sm:rounded sm:text-xs sm:leading-tight">{t('dashboard.proteinsConsumed')}: {dailyTotals.proteins.toFixed(1)}g</span>
                        )}
                      </div>
                    )}
                  </div>
                  
                  {/* Entries list - only visible when expanded */}
                  {isExpanded && (
                    <>
                      <ul className="list-none p-0 m-0">
                        {entriesByDate[dateKey].map((entry: any) => (
                          <EntryListItem key={entry.id} entry={entry} onEdit={onEdit} />
                        ))}
                      </ul>

                      {/* Daily totals footer - only when expanded */}
                      <div className="mt-4 p-3 bg-gray-50 rounded border border-gray-300">
                        <div className="mt-4 p-3 bg-gray-50 rounded border border-gray-300">
                          <div className="flex items-center justify-between font-bold text-sm mb-2">
                            <span className="text-gray-700">{t('dashboard.total')}:</span>
                            <span className="text-green-600">{dailyTotals.calories} {t('dashboard.kcal')}</span>
                          </div>
                          {(dailyTotals.fats > 0 || dailyTotals.carbs > 0 || dailyTotals.proteins > 0) && (
                            <div className="flex flex-wrap gap-2 text-sm md:gap-2 sm:gap-1">
                              {dailyTotals.fats > 0 && (
                                <span className="flex-shrink-0 md:bg-transparent md:p-0 md:rounded-none md:text-inherit md:leading-normal sm:flex-[0_0_calc(50%-0.25rem)] sm:text-center sm:p-1 sm:bg-green-100 sm:rounded sm:text-xs sm:leading-tight">{t('dashboard.fatsConsumed')}: {dailyTotals.fats.toFixed(1)}g</span>
                              )}
                              {dailyTotals.carbs > 0 && (
                                <span className="flex-shrink-0 md:bg-transparent md:p-0 md:rounded-none md:text-inherit md:leading-normal sm:flex-[0_0_calc(50%-0.25rem)] sm:text-center sm:p-1 sm:bg-green-100 sm:rounded sm:text-xs sm:leading-tight">{t('dashboard.carbsConsumed')}: {dailyTotals.carbs.toFixed(1)}g</span>
                              )}
                              {dailyTotals.proteins > 0 && (
                                <span className="flex-shrink-0 md:bg-transparent md:p-0 md:rounded-none md:text-inherit md:leading-normal sm:flex-[0_0_calc(50%-0.25rem)] sm:text-center sm:p-1 sm:bg-green-100 sm:rounded sm:text-xs sm:leading-tight">{t('dashboard.proteinsConsumed')}: {dailyTotals.proteins.toFixed(1)}g</span>
                              )}
                            </div>
                          )}
                        </div>
                      </div>
                    </>
                  )}
                </div>
              );
            })}
        </div>
      )}
    </section>
  );
}