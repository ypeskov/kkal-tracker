import { useTranslation } from 'react-i18next';

export default function Report() {
  const { t } = useTranslation();

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">{t('report.title')}</h2>
      </div>
      <div className="card p-lg">
        <p>{t('report.wip')}</p>
      </div>
    </div>
  );
}