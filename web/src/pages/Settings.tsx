import { useTranslation } from 'react-i18next';

export default function Settings() {
  const { t } = useTranslation();

  return (
    <div>
      <h2>{t('settings.title')}</h2>
      <p>{t('settings.wip')}</p>
    </div>
  );
}