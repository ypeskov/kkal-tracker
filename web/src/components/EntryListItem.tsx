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
  
  // Calculate actual nutrient amounts consumed
  const actualFats = entry.fats ? (entry.fats * entry.weight / 100) : null;
  const actualCarbs = entry.carbs ? (entry.carbs * entry.weight / 100) : null;
  const actualProteins = entry.proteins ? (entry.proteins * entry.weight / 100) : null;

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
        {actualFats !== null ? (
          <span>{t('dashboard.fatsConsumed')}: {actualFats.toFixed(1)}g</span>
        ) : null}
        {actualCarbs !== null ? (
          <span>{t('dashboard.carbsConsumed')}: {actualCarbs.toFixed(1)}g</span>
        ) : null}
        {actualProteins !== null ? (
          <span>{t('dashboard.proteinsConsumed')}: {actualProteins.toFixed(1)}g</span>
        ) : null}
      </div>
    </li>
  );
}