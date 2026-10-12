import TabNavigation from '@/components/TabNavigation';
import AiTab from '@/components/settings/AiTab';
import ApiKeysTab from '@/components/settings/ApiKeysTab';
import ExportTab from '@/components/settings/ExportTab';
import { Bot, Download, Key } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

type SettingsTab = 'export' | 'apiKeys' | 'ai';

export default function Settings() {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<SettingsTab>('export');

  return (
    <div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">
      <div className="mb-6">
        <h2 className="text-3xl font-semibold text-gray-800">{t('settings.title')}</h2>
      </div>

      <TabNavigation
        tabs={[
          { id: 'export', label: t('settings.export.tab'), icon: <Download size={18} /> },
          { id: 'apiKeys', label: t('settings.apiKeys.tab'), icon: <Key size={18} /> },
          { id: 'ai', label: t('settings.ai.tab'), icon: <Bot size={18} /> },
        ]}
        activeTab={activeTab}
        onTabChange={(tabId) => setActiveTab(tabId as SettingsTab)}
      />

      {activeTab === 'export' && <ExportTab />}
      {activeTab === 'apiKeys' && <ApiKeysTab />}
      {activeTab === 'ai' && <AiTab />}
    </div>
  );
}
