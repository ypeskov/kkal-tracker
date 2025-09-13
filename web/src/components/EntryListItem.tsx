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
    <li className="bg-white p-3 mb-2 rounded-md border border-gray-200 shadow-sm transition-all duration-200 cursor-pointer hover:bg-gray-50 hover:border-blue-400 hover:shadow-lg hover:-translate-y-0.5" onClick={() => onEdit(entry)}>
      <div className="flex items-center w-full mb-2">
        <span className="text-sm text-gray-500 w-[15%] min-w-[45px] font-normal flex-shrink-0">
          {formatTime(entry.meal_datetime)}
        </span>
        <strong className="font-semibold w-[60%] overflow-hidden text-ellipsis whitespace-nowrap pr-2 text-gray-800">{entry.food}</strong>
        <strong className="font-semibold w-[25%] text-right flex-shrink-0 whitespace-nowrap text-gray-800">{entry.calories} {t('dashboard.kcal')}</strong>
      </div>
      <div className="flex flex-wrap gap-2 text-base text-gray-600 md:flex-nowrap md:gap-4 sm:gap-1">
        <span className="flex-shrink-0">{t('dashboard.weight')}: {entry.weight}g</span>
        {(entry.fats !== null && entry.fats !== undefined) && (
          <span className="flex-shrink-0">{t('dashboard.fatsConsumed')}: {actualFats?.toFixed(1)}g</span>
        )}
        {(entry.carbs !== null && entry.carbs !== undefined) && (
          <span className="flex-shrink-0">{t('dashboard.carbsConsumed')}: {actualCarbs?.toFixed(1)}g</span>
        )}
        {(entry.proteins !== null && entry.proteins !== undefined) && (
          <span className="flex-shrink-0">{t('dashboard.proteinsConsumed')}: {actualProteins?.toFixed(1)}g</span>
        )}
      </div>
    </li>
  );
}