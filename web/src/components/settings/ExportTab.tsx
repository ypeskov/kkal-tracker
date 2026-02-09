import { exportService, ExportDataType, DeliveryType } from '@/api/export';
import { profileService } from '@/api/profile';
import NotificationPopup from '@/components/NotificationPopup';
import { useMutation, useQuery } from '@tanstack/react-query';
import { format, subDays } from 'date-fns';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

export default function ExportTab() {
  const { t } = useTranslation();

  // Default: last 30 days
  const [dateFrom, setDateFrom] = useState(format(subDays(new Date(), 30), 'yyyy-MM-dd'));
  const [dateTo, setDateTo] = useState(format(new Date(), 'yyyy-MM-dd'));
  const [dataType, setDataType] = useState<ExportDataType>('both');
  const [deliveryType, setDeliveryType] = useState<DeliveryType>('download');
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // Get user profile for email
  const { data: profile } = useQuery({
    queryKey: ['profile'],
    queryFn: profileService.getProfile,
  });

  const exportMutation = useMutation({
    mutationFn: () =>
      exportService.exportData({
        date_from: dateFrom,
        date_to: dateTo,
        data_type: dataType,
        delivery_type: deliveryType,
      }),
    onSuccess: (data) => {
      if (data instanceof Blob) {
        const filename = `kkal-export-${dateFrom}-to-${dateTo}.xlsx`;
        exportService.downloadBlob(data, filename);
        setNotification({ type: 'success', message: t('settings.export.downloadSuccess') });
      } else {
        setNotification({ type: 'success', message: t('settings.export.emailSuccess') });
      }
    },
    onError: (error: Error) => {
      setNotification({ type: 'error', message: error.message || t('settings.export.error') });
    },
  });

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h3 className="text-xl font-semibold text-gray-800 mb-4">{t('settings.export.title')}</h3>

      <div className="space-y-6">
        {/* Date Range */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('settings.export.dateFrom')}
            </label>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('settings.export.dateTo')}
            </label>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        {/* Data Type Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            {t('settings.export.dataType')}
          </label>
          <div className="flex flex-wrap gap-4">
            {(['weight', 'food', 'both'] as ExportDataType[]).map((type) => (
              <label key={type} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="dataType"
                  value={type}
                  checked={dataType === type}
                  onChange={() => setDataType(type)}
                  className="w-4 h-4 text-blue-600"
                />
                <span>{t(`settings.export.dataTypes.${type}`)}</span>
              </label>
            ))}
          </div>
        </div>

        {/* Delivery Method */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            {t('settings.export.deliveryMethod')}
          </label>
          <div className="flex flex-wrap gap-4">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="deliveryType"
                value="download"
                checked={deliveryType === 'download'}
                onChange={() => setDeliveryType('download')}
                className="w-4 h-4 text-blue-600"
              />
              <span>{t('settings.export.download')}</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                name="deliveryType"
                value="email"
                checked={deliveryType === 'email'}
                onChange={() => setDeliveryType('email')}
                className="w-4 h-4 text-blue-600"
              />
              <span>{t('settings.export.email')}</span>
              {profile?.email && (
                <span className="text-gray-500 text-sm">({profile.email})</span>
              )}
            </label>
          </div>
        </div>

        {/* Export Button */}
        <div className="flex justify-end">
          <button
            onClick={() => exportMutation.mutate()}
            disabled={exportMutation.isPending}
            className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {exportMutation.isPending ? t('settings.export.exporting') : t('settings.export.exportButton')}
          </button>
        </div>
      </div>

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
