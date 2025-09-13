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
    <section>
      <h2>{t('dashboard.todayEntries')}</h2>
      {entries?.length === 0 ? (
        <p>{t('dashboard.noEntries')}</p>
      ) : (
        <div className="entries-list">
          {Object.keys(entriesByDate)
            .sort((a, b) => new Date(b).getTime() - new Date(a).getTime()) // Sort dates desc (newest first)
            .map((dateKey, index) => {
              const isExpanded = !hasMultipleDates || expandedDates.has(dateKey);
              const dailyTotals = calculateDailyTotals(entriesByDate[dateKey]);
              
              return (
                <div key={dateKey} className={`date-section ${
                  index % 2 === 0 ? 'date-section--primary' : 'date-section--secondary'
                }`}>
                  {/* Clickable date header with total */}
                  <div
                    onClick={() => hasMultipleDates && toggleDateExpansion(dateKey)}
                    className={`date-header ${
                      index % 2 === 0 ? 'date-header--primary' : 'date-header--secondary'
                    } ${
                      hasMultipleDates ? 'date-header--clickable' : ''
                    } ${
                      isExpanded ? 'date-header--expanded' : 'date-header--collapsed'
                    }`}
                  >
                    <div className="date-header__content">
                      <div className="date-header__title">
                        {hasMultipleDates && (
                          <span className={`date-header__toggle ${
                            isExpanded ? 'date-header__toggle--expanded' : 'date-header__toggle--collapsed'
                          }`}>
                            ▼
                          </span>
                        )}
                        <span>{formatDateHeader(dateKey)}</span>
                      </div>
                      <span className="date-header__calories">
                        {dailyTotals.calories} {t('dashboard.kcal')}
                      </span>
                    </div>
                    
                    {/* Show nutrient summary in header when collapsed */}
                    {!isExpanded && (dailyTotals.fats > 0 || dailyTotals.carbs > 0 || dailyTotals.proteins > 0) && (
                      <div className="nutrients-summary">
                        {dailyTotals.fats > 0 && (
                          <span className="nutrients-summary__item">{t('dashboard.fatsConsumed')}: {dailyTotals.fats.toFixed(1)}g</span>
                        )}
                        {dailyTotals.carbs > 0 && (
                          <span className="nutrients-summary__item">{t('dashboard.carbsConsumed')}: {dailyTotals.carbs.toFixed(1)}g</span>
                        )}
                        {dailyTotals.proteins > 0 && (
                          <span className="nutrients-summary__item">{t('dashboard.proteinsConsumed')}: {dailyTotals.proteins.toFixed(1)}g</span>
                        )}
                      </div>
                    )}
                  </div>
                  
                  {/* Entries list - only visible when expanded */}
                  {isExpanded && (
                    <>
                      <ul className="entries-list__items">
                        {entriesByDate[dateKey].map((entry: any) => (
                          <EntryListItem key={entry.id} entry={entry} onEdit={onEdit} />
                        ))}
                      </ul>

                      {/* Daily totals footer - only when expanded */}
                      <div className="daily-totals">
                        <div className="daily-totals">
                          <div className="daily-totals__header">
                            <span className="daily-totals__label">{t('dashboard.total')}:</span>
                            <span className="daily-totals__calories">{dailyTotals.calories} {t('dashboard.kcal')}</span>
                          </div>
                          {(dailyTotals.fats > 0 || dailyTotals.carbs > 0 || dailyTotals.proteins > 0) && (
                            <div className="daily-totals__nutrients">
                              {dailyTotals.fats > 0 && (
                                <span className="daily-totals__nutrient">{t('dashboard.fatsConsumed')}: {dailyTotals.fats.toFixed(1)}g</span>
                              )}
                              {dailyTotals.carbs > 0 && (
                                <span className="daily-totals__nutrient">{t('dashboard.carbsConsumed')}: {dailyTotals.carbs.toFixed(1)}g</span>
                              )}
                              {dailyTotals.proteins > 0 && (
                                <span className="daily-totals__nutrient">{t('dashboard.proteinsConsumed')}: {dailyTotals.proteins.toFixed(1)}g</span>
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