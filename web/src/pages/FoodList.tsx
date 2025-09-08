import { useTranslation } from 'react-i18next';

export default function FoodList() {
  const { t } = useTranslation();

  return (
    <div>
      <h2>{t('foodList.title')}</h2>
      <p>{t('foodList.wip')}</p>
    </div>
  );
}