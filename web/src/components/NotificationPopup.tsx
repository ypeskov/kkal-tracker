import { useTranslation } from 'react-i18next';

interface NotificationPopupProps {
  type: 'success' | 'error';
  message: string;
  onClose: () => void;
}

export default function NotificationPopup({ type, message, onClose }: NotificationPopupProps) {
  const { t } = useTranslation();

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn px-4">
      <div className="bg-white rounded-lg w-full md:w-[600px] lg:w-[700px] p-8 shadow-xl animate-slideUp text-center">
        <h3 className={`mt-0 mb-4 text-lg font-medium ${
          type === 'success' ? 'text-green-600' : 'text-red-600'
        }`}>
          {type === 'success' ? t('common.success') : t('common.error')}
        </h3>
        <p className="mb-6 text-gray-700">{message}</p>
        <div className="flex justify-center">
          <button
            onClick={onClose}
            className="btn-primary px-4 py-2 text-sm font-medium"
          >
            {t('common.ok')}
          </button>
        </div>
      </div>
    </div>
  );
}
