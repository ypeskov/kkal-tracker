import TabNavigation from '@/components/TabNavigation';
import ExportTab from '@/components/settings/ExportTab';
import { Download } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

export default function Settings() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<'export'>('export');

  return (
    <div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">
      <div className="mb-6">
        <h2 className="text-3xl font-semibold text-gray-800">{t('settings.title')}</h2>
      </div>

      <TabNavigation
        tabs={[
          { id: 'export', label: t('settings.export.tab'), icon: <Download size={18} /> },
          // Future: { id: 'apiKeys', label: t('settings.apiKeys.tab'), icon: <Key size={18} /> },
        ]}
        activeTab={activeTab}
        onTabChange={(tabId) => setActiveTab(tabId as 'export')}
      />

      {activeTab === 'export' && <ExportTab />}
    </div>
  );
}
