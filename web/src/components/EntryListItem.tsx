import { useTranslation } from 'react-i18next';

interface EntryListItemProps {
  entry: any;
  onEdit: (entry: any) => void;
}

export default function EntryListItem({ entry, onEdit }: EntryListItemProps) {
  const { t, i18n } = useTranslation();

  const formatTime = (datetimeString: string) => {
    const date = new Date(datetimeString);
    const currentLang = i18n.language;
    const locale = currentLang === 'uk-UA' ? 'uk-UA' : currentLang === 'ru-UA' ? 'ru-RU' : 'en-US';
    return date.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <li className="entry-item" onClick={() => onEdit(entry)} style={{ cursor: 'pointer' }}>
      <div className="entry-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <span style={{ 
            fontSize: '0.85rem', 
            color: '#6c757d',
            minWidth: '55px',
            fontWeight: 'normal'
          }}>
            {formatTime(entry.meal_datetime)}
          </span>
          <strong>{entry.food}</strong>
        </div>
        <strong>{entry.calories} {t('dashboard.kcal')}</strong>
      </div>
      <div className="entry-details" style={{ marginLeft: '60px' }}>
        <span>{t('dashboard.weight')}: {entry.weight}g</span>
        <span>{t('dashboard.kcalPer100g')}: {entry.kcalPer100g}</span>
        {entry.fats && <span>{t('dashboard.fats')}: {entry.fats}g</span>}
        {entry.carbs && <span>{t('dashboard.carbs')}: {entry.carbs}g</span>}
        {entry.proteins && <span>{t('dashboard.proteins')}: {entry.proteins}g</span>}
      </div>
    </li>
  );
}