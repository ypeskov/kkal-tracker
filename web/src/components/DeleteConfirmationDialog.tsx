import { useTranslation } from 'react-i18next';

interface DeleteConfirmationDialogProps {
  title?: string;
  message: string;
  itemName?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isDeleting: boolean;
}

export default function DeleteConfirmationDialog({ 
  title,
  message,
  itemName,
  onConfirm, 
  onCancel, 
  isDeleting 
}: DeleteConfirmationDialogProps) {
  const { t } = useTranslation();

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn">
      <div className="bg-white rounded-lg max-w-sm w-full max-w-[90%] p-6 shadow-xl animate-slideUp text-center">
        <h3 className="text-red-600 mt-0 mb-4 text-lg font-medium">
          {title || t('dashboard.confirmDelete')}
        </h3>
        <p className="mb-6 text-gray-700">
          {itemName ? t(message, { foodName: itemName }) : message}
        </p>
        
        <div className="flex justify-center gap-3">
          <button
            type="button"
            onClick={onCancel}
            className="btn-secondary px-4 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed"
            disabled={isDeleting}
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="btn-danger px-4 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed"
            disabled={isDeleting}
          >
            {isDeleting ? t('common.deleting') : t('common.delete')}
          </button>
        </div>
      </div>
    </div>
  );
}