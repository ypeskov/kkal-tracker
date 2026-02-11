import { apiKeysService, CreateAPIKeyResponse } from '@/api/apikeys';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import NotificationPopup from '@/components/NotificationPopup';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { format } from 'date-fns';
import { Copy, Check } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

type ExpiryOption = 'never' | '30' | '90' | '180' | '365' | 'custom';

export default function ApiKeysTab() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [name, setName] = useState('');
  const [expiryOption, setExpiryOption] = useState<ExpiryOption>('never');
  const [customDays, setCustomDays] = useState('');
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [createdKey, setCreatedKey] = useState<CreateAPIKeyResponse | null>(null);
  const [copied, setCopied] = useState(false);
  const [deleteKeyId, setDeleteKeyId] = useState<number | null>(null);

  const { data: keys = [], isLoading } = useQuery({
    queryKey: ['apiKeys'],
    queryFn: apiKeysService.listKeys,
  });

  const createMutation = useMutation({
    mutationFn: () => {
      let expiryDays: number | undefined;
      if (expiryOption !== 'never') {
        expiryDays = expiryOption === 'custom' ? parseInt(customDays, 10) : parseInt(expiryOption, 10);
      }
      return apiKeysService.createKey({ name, expiry_days: expiryDays });
    },
    onSuccess: (data) => {
      setCreatedKey(data);
      setName('');
      setExpiryOption('never');
      setCustomDays('');
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
    },
    onError: () => {
      setNotification({ type: 'error', message: t('settings.apiKeys.error') });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => apiKeysService.deleteKey(id),
    onSuccess: () => {
      setNotification({ type: 'success', message: t('settings.apiKeys.deleteSuccess') });
      setDeleteKeyId(null);
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
    },
    onError: () => {
      setNotification({ type: 'error', message: t('settings.apiKeys.error') });
      setDeleteKeyId(null);
    },
  });

  const handleCopy = async (key: string) => {
    await navigator.clipboard.writeText(key);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const canCreate = name.trim() && (expiryOption !== 'custom' || (customDays && parseInt(customDays, 10) > 0));


  return (
    <div className="space-y-6">
      {/* Create Key Form */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <h3 className="text-xl font-semibold text-gray-800 mb-2">{t('settings.apiKeys.title')}</h3>
        <p className="text-gray-600 mb-4">{t('settings.apiKeys.description')}</p>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('settings.apiKeys.name')}
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('settings.apiKeys.namePlaceholder')}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              maxLength={100}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {t('settings.apiKeys.expiry')}
            </label>
            <div className="flex flex-wrap gap-3">
              {(['never', '30', '90', '180', '365', 'custom'] as ExpiryOption[]).map((option) => (
                <label key={option} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    name="expiry"
                    value={option}
                    checked={expiryOption === option}
                    onChange={() => setExpiryOption(option)}
                    className="w-4 h-4 text-blue-600"
                  />
                  <span className="text-sm">{t(`settings.apiKeys.expiryOptions.${option}`)}</span>
                </label>
              ))}
            </div>
            {expiryOption === 'custom' && (
              <input
                type="number"
                value={customDays}
                onChange={(e) => setCustomDays(e.target.value)}
                placeholder={t('settings.apiKeys.customDays')}
                min="1"
                max="3650"
                className="mt-2 w-40 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            )}
          </div>

          <div className="flex justify-end">
            <button
              onClick={() => createMutation.mutate()}
              disabled={!canCreate || createMutation.isPending}
              className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
            >
              {createMutation.isPending ? t('settings.apiKeys.creating') : t('settings.apiKeys.create')}
            </button>
          </div>
        </div>
      </div>

      {/* Key List */}
      <div className="bg-white rounded-lg shadow-md p-6">
        {isLoading ? (
          <p className="text-gray-500">{t('common.loading')}</p>
        ) : keys.length === 0 ? (
          <p className="text-gray-500">{t('settings.apiKeys.noKeys')}</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="pb-3 text-sm font-medium text-gray-600">{t('settings.apiKeys.columns.name')}</th>
                  <th className="pb-3 text-sm font-medium text-gray-600">{t('settings.apiKeys.columns.key')}</th>
                  <th className="pb-3 text-sm font-medium text-gray-600">{t('settings.apiKeys.columns.created')}</th>
                  <th className="pb-3 text-sm font-medium text-gray-600">{t('settings.apiKeys.columns.expires')}</th>
                  <th className="pb-3 text-sm font-medium text-gray-600">{t('settings.apiKeys.columns.actions')}</th>
                </tr>
              </thead>
              <tbody>
                {keys.map((key) => (
                    <tr key={key.id} className="border-b border-gray-100">
                      <td className="py-3 text-sm">{key.name}</td>
                      <td className="py-3 text-sm font-mono text-gray-500">{key.key_prefix}...</td>
                      <td className="py-3 text-sm text-gray-500">
                        {format(new Date(key.created_at), 'yyyy-MM-dd')}
                      </td>
                      <td className="py-3 text-sm text-gray-500">
                        {key.expires_at
                          ? format(new Date(key.expires_at), 'yyyy-MM-dd')
                          : t('settings.apiKeys.never')}
                      </td>
                      <td className="py-3">
                        <button
                          onClick={() => setDeleteKeyId(key.id)}
                          className="text-sm text-red-600 hover:text-red-700"
                        >
                          {t('settings.apiKeys.delete')}
                        </button>
                      </td>
                    </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Usage Hint */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <h4 className="text-lg font-semibold text-gray-800 mb-3">{t('settings.apiKeys.usage.title')}</h4>
        <div className="space-y-2 text-sm">
          <div>
            <span className="font-medium text-gray-700">{t('settings.apiKeys.usage.endpoint')}:</span>{' '}
            <code className="bg-gray-100 px-2 py-0.5 rounded text-gray-800">GET /api/v1/data</code>
          </div>
          <div>
            <span className="font-medium text-gray-700">{t('settings.apiKeys.usage.header')}:</span>{' '}
            <code className="bg-gray-100 px-2 py-0.5 rounded text-gray-800">X-API-Key: your-key</code>
          </div>
          <div>
            <span className="font-medium text-gray-700">{t('settings.apiKeys.usage.params')}:</span>
            <ul className="ml-4 mt-1 list-disc text-gray-600">
              <li>{t('settings.apiKeys.usage.paramType')}</li>
              <li>{t('settings.apiKeys.usage.paramFrom')}</li>
              <li>{t('settings.apiKeys.usage.paramTo')}</li>
            </ul>
          </div>
        </div>
      </div>

      {/* Created Key Modal */}
      {createdKey && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn px-4">
          <div className="bg-white rounded-lg w-full md:w-[600px] lg:w-[700px] p-8 shadow-xl animate-slideUp">
            <h3 className="text-lg font-semibold text-gray-800 mb-2">{t('settings.apiKeys.created.title')}</h3>
            <p className="text-gray-600 mb-1">{t('settings.apiKeys.created.message')}</p>
            <p className="text-amber-600 text-sm mb-4">{t('settings.apiKeys.created.warning')}</p>

            <div className="flex items-center gap-2 mb-4">
              <input
                type="text"
                readOnly
                value={createdKey.key}
                className="flex-1 min-w-0 px-3 py-2 text-sm font-mono bg-gray-50 border border-gray-200 rounded-md"
                onClick={(e) => (e.target as HTMLInputElement).select()}
              />
              <button
                onClick={() => handleCopy(createdKey.key)}
                className="flex-shrink-0 p-2 text-gray-500 hover:text-blue-600 transition-colors"
                title={t('settings.apiKeys.created.copy')}
              >
                {copied ? <Check size={18} className="text-green-600" /> : <Copy size={18} />}
              </button>
            </div>

            {copied && (
              <p className="text-green-600 text-sm mb-3">{t('settings.apiKeys.created.copied')}</p>
            )}

            <div className="flex justify-end">
              <button
                onClick={() => setCreatedKey(null)}
                className="btn-primary px-4 py-2 text-sm font-medium"
              >
                {t('settings.apiKeys.created.close')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Confirm Dialog */}
      {deleteKeyId !== null && (
        <DeleteConfirmationDialog
          title={t('settings.apiKeys.delete')}
          message={t('settings.apiKeys.deleteConfirm')}
          onConfirm={() => deleteMutation.mutate(deleteKeyId)}
          onCancel={() => setDeleteKeyId(null)}
          isDeleting={deleteMutation.isPending}
        />
      )}

      {notification && (
        <NotificationPopup
          type={notification.type}
          message={notification.message}
          onClose={() => setNotification(null)}
        />
      )}
    </div>
  );
}
