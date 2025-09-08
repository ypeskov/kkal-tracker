import { useTranslation } from 'react-i18next';

export default function Report() {
  const { t } = useTranslation();

  return (
    <div>
      <h2>{t('report.title')}</h2>
      <p>{t('report.wip')}</p>
    </div>
  );
}