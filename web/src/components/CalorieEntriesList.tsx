import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import EntryListItem from './EntryListItem';

interface CalorieEntriesListProps {
  entries: any[];
  onEdit: (entry: any) => void;
  filterType: string;
}

export default function CalorieEntriesList({ entries, onEdit, filterType }: CalorieEntriesListProps) {
  const { t, i18n } = useTranslation();

  const entriesByDate = useMemo(() => {
    if (!entries || entries.length === 0) {
      return {};
    }

    const grouped: { [date: string]: any[] } = {};

    entries.forEach(entry => {
      const date = new Date(entry.meal_datetime).toISOString().split('T')[0];
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

    const dateOnly = date.toISOString().split('T')[0];
    const todayOnly = today.toISOString().split('T')[0];
    const yesterdayOnly = yesterday.toISOString().split('T')[0];

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

  return (
    <section>
      <h2>{t('dashboard.todayEntries')}</h2>
      {entries?.length === 0 ? (
        <p>{t('dashboard.noEntries')}</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
          {Object.keys(entriesByDate)
            .sort((a, b) => new Date(b).getTime() - new Date(a).getTime()) // Sort dates desc (newest first)
            .map((dateKey) => (
              <div key={dateKey}>
                {(filterType !== 'today' || Object.keys(entriesByDate).length > 1) && (
                  <h3 style={{ 
                    margin: '0 0 1rem 0', 
                    padding: '0.5rem 0', 
                    fontSize: '1.1rem',
                    color: '#495057',
                    borderBottom: '2px solid #e9ecef'
                  }}>
                    {formatDateHeader(dateKey)}
                  </h3>
                )}
                
                <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                  {entriesByDate[dateKey].map((entry: any) => (
                    <EntryListItem key={entry.id} entry={entry} onEdit={onEdit} />
                  ))}
                </ul>

                {Object.keys(entriesByDate).length > 1 && (() => {
                  const dailyTotals = calculateDailyTotals(entriesByDate[dateKey]);
                  return (
                    <div style={{ 
                      marginTop: '1rem', 
                      padding: '0.75rem', 
                      backgroundColor: '#f8f9fa', 
                      borderRadius: '5px',
                      border: '1px solid #dee2e6'
                    }}>
                      <div style={{ 
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: '1rem',
                        fontWeight: 'bold',
                        fontSize: '0.9rem'
                      }}>
                        <span style={{ color: '#495057' }}>{t('dashboard.total')}:</span>
                        <span>{dailyTotals.calories} {t('dashboard.kcal')}</span>
                        {dailyTotals.fats > 0 && (
                          <span>{t('dashboard.fats')}: {dailyTotals.fats.toFixed(1)}g</span>
                        )}
                        {dailyTotals.carbs > 0 && (
                          <span>{t('dashboard.carbs')}: {dailyTotals.carbs.toFixed(1)}g</span>
                        )}
                        {dailyTotals.proteins > 0 && (
                          <span>{t('dashboard.proteins')}: {dailyTotals.proteins.toFixed(1)}g</span>
                        )}
                      </div>
                    </div>
                  )
                })()}
              </div>
            ))}
        </div>
      )}
    </section>
  );
}