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
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          {Object.keys(entriesByDate)
            .sort((a, b) => new Date(b).getTime() - new Date(a).getTime()) // Sort dates desc (newest first)
            .map((dateKey, index) => {
              const isExpanded = !hasMultipleDates || expandedDates.has(dateKey);
              const dailyTotals = calculateDailyTotals(entriesByDate[dateKey]);
              
              return (
                <div key={dateKey} className="date-section" style={{
                  background: index % 2 === 0 ? '#f0f4f8' : '#e8f5e9',
                  borderRadius: '8px',
                  padding: '0.75rem',
                  marginBottom: '1.5rem',
                  border: '1px solid ' + (index % 2 === 0 ? '#d1dae3' : '#c3e6c8'),
                  boxShadow: '0 2px 6px rgba(0,0,0,0.08)'
                }}>
                  {/* Clickable date header with total */}
                  <div 
                    onClick={() => hasMultipleDates && toggleDateExpansion(dateKey)}
                    style={{ 
                      cursor: hasMultipleDates ? 'pointer' : 'default',
                      margin: isExpanded ? '0 0 1rem 0' : '0', 
                      padding: '0.75rem 1rem', 
                      fontSize: '1.1rem',
                      color: index % 2 === 0 ? '#1a5490' : '#2d6a4f',
                      fontWeight: '600',
                      background: index % 2 === 0 ? 'linear-gradient(90deg, #ffffff, #f8fbff)' : 'linear-gradient(90deg, #ffffff, #f0fdf4)',
                      borderRadius: '6px',
                      borderLeft: '5px solid ' + (index % 2 === 0 ? '#4a90e2' : '#52c41a'),
                      boxShadow: '0 2px 8px ' + (index % 2 === 0 ? 'rgba(74, 144, 226, 0.15)' : 'rgba(82, 196, 26, 0.15)'),
                      transition: 'transform 0.2s ease',
                      ...(hasMultipleDates && {
                        ':hover': {
                          transform: 'translateX(2px)'
                        }
                      })
                    }}
                    className="date-header-clickable"
                  >
                    <div style={{ 
                      display: 'flex', 
                      alignItems: 'center',
                      justifyContent: 'space-between'
                    }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        {hasMultipleDates && (
                          <span style={{
                            display: 'inline-block',
                            width: '20px',
                            height: '20px',
                            lineHeight: '20px',
                            textAlign: 'center',
                            fontSize: '14px',
                            fontWeight: 'bold',
                            transition: 'transform 0.2s ease',
                            transform: isExpanded ? 'rotate(0deg)' : 'rotate(-90deg)'
                          }}>
                            ▼
                          </span>
                        )}
                        <span>{formatDateHeader(dateKey)}</span>
                      </div>
                      <span style={{ 
                        color: '#28a745',
                        fontWeight: 'bold',
                        fontSize: '1rem'
                      }}>
                        {dailyTotals.calories} {t('dashboard.kcal')}
                      </span>
                    </div>
                    
                    {/* Show nutrient summary in header when collapsed */}
                    {!isExpanded && (dailyTotals.fats > 0 || dailyTotals.carbs > 0 || dailyTotals.proteins > 0) && (
                      <div style={{
                        display: 'flex',
                        flexWrap: 'wrap',
                        gap: '0.75rem',
                        fontSize: '0.85rem',
                        marginTop: '0.5rem',
                        opacity: 0.8
                      }}>
                        {dailyTotals.fats > 0 && (
                          <span>{t('dashboard.fatsConsumed')}: {dailyTotals.fats.toFixed(1)}g</span>
                        )}
                        {dailyTotals.carbs > 0 && (
                          <span>{t('dashboard.carbsConsumed')}: {dailyTotals.carbs.toFixed(1)}g</span>
                        )}
                        {dailyTotals.proteins > 0 && (
                          <span>{t('dashboard.proteinsConsumed')}: {dailyTotals.proteins.toFixed(1)}g</span>
                        )}
                      </div>
                    )}
                  </div>
                  
                  {/* Entries list - only visible when expanded */}
                  {isExpanded && (
                    <>
                      <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                        {entriesByDate[dateKey].map((entry: any) => (
                          <EntryListItem key={entry.id} entry={entry} onEdit={onEdit} />
                        ))}
                      </ul>

                      {/* Daily totals footer - only when expanded */}
                      <div style={{ 
                        marginTop: '1rem', 
                        padding: '0.75rem', 
                        backgroundColor: '#f8f9fa', 
                        borderRadius: '5px',
                        border: '1px solid #dee2e6'
                      }}>
                        <div className="daily-totals">
                          <div style={{ 
                            display: 'flex', 
                            alignItems: 'center',
                            justifyContent: 'space-between',
                            fontWeight: 'bold',
                            fontSize: '0.95rem',
                            marginBottom: '0.5rem'
                          }}>
                            <span style={{ color: '#495057' }}>{t('dashboard.total')}:</span>
                            <span style={{ color: '#28a745' }}>{dailyTotals.calories} {t('dashboard.kcal')}</span>
                          </div>
                          {(dailyTotals.fats > 0 || dailyTotals.carbs > 0 || dailyTotals.proteins > 0) && (
                            <div className="daily-totals-nutrients" style={{
                              display: 'flex',
                              flexWrap: 'wrap',
                              gap: '0.5rem',
                              fontSize: '0.9rem'
                            }}>
                              {dailyTotals.fats > 0 && (
                                <span>{t('dashboard.fatsConsumed')}: {dailyTotals.fats.toFixed(1)}g</span>
                              )}
                              {dailyTotals.carbs > 0 && (
                                <span>{t('dashboard.carbsConsumed')}: {dailyTotals.carbs.toFixed(1)}g</span>
                              )}
                              {dailyTotals.proteins > 0 && (
                                <span>{t('dashboard.proteinsConsumed')}: {dailyTotals.proteins.toFixed(1)}g</span>
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