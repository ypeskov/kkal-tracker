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
    <li className="entry-item" onClick={() => onEdit(entry)}>
      <div className="entry-item__header">
        <span className="entry-item__time">
          {formatTime(entry.meal_datetime)}
        </span>
        <strong className="entry-item__food">{entry.food}</strong>
        <strong className="entry-item__calories">{entry.calories} {t('dashboard.kcal')}</strong>
      </div>
      <div className="entry-item__details">
        <span className="entry-item__detail">{t('dashboard.weight')}: {entry.weight}g</span>
        {actualFats !== null ? (
          <span className="entry-item__detail">{t('dashboard.fatsConsumed')}: {actualFats.toFixed(1)}g</span>
        ) : null}
        {actualCarbs !== null ? (
          <span className="entry-item__detail">{t('dashboard.carbsConsumed')}: {actualCarbs.toFixed(1)}g</span>
        ) : null}
        {actualProteins !== null ? (
          <span className="entry-item__detail">{t('dashboard.proteinsConsumed')}: {actualProteins.toFixed(1)}g</span>
        ) : null}
      </div>
    </li>
  );
}