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
      <div className="entry-header" style={{ 
        display: 'flex',
        alignItems: 'center',
        width: '100%'
      }}>
        <span style={{ 
          fontSize: '0.85rem', 
          color: '#6c757d',
          width: '15%',
          minWidth: '45px',
          fontWeight: 'normal',
          flexShrink: 0
        }}>
          {formatTime(entry.meal_datetime)}
        </span>
        <strong style={{ 
          width: '60%',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          paddingRight: '0.5rem'
        }}>{entry.food}</strong>
        <strong style={{
          width: '25%',
          textAlign: 'right',
          flexShrink: 0,
          whiteSpace: 'nowrap'
        }}>{entry.calories} {t('dashboard.kcal')}</strong>
      </div>
      <div className="entry-details">
        <span>{t('dashboard.weight')}: {entry.weight}g</span>
        <span>{t('dashboard.kcalPer100g')}: {entry.kcalPer100g}</span>
        {entry.fats && <span>{t('dashboard.fats')}: {entry.fats}g/100g</span>}
        {entry.carbs && <span>{t('dashboard.carbs')}: {entry.carbs}g/100g</span>}
        {entry.proteins && <span>{t('dashboard.proteins')}: {entry.proteins}g/100g</span>}
      </div>
    </li>
  );
}