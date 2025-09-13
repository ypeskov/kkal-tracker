import { useTranslation } from 'react-i18next';

export default function Settings() {
  const { t } = useTranslation();

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">{t('settings.title')}</h2>
      </div>
      <div className="card p-lg">
        <p>{t('settings.wip')}</p>
      </div>
    </div>
  );
}