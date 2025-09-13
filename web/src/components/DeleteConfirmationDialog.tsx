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
    <div className="modal-overlay z-modal-high">
      <div className="modal-content text-center max-w-sm">
        <h3 className="text-error mt-0">
          {title || t('dashboard.confirmDelete')}
        </h3>
        <p className="mb-2xl">
          {itemName ? t(message, { foodName: itemName }) : message}
        </p>
        
        <div className="flex justify-center gap-lg">
          <button
            type="button"
            onClick={onCancel}
            className="btn btn--secondary"
            disabled={isDeleting}
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="btn btn--danger"
            disabled={isDeleting}
          >
            {isDeleting ? t('common.deleting') : t('common.delete')}
          </button>
        </div>
      </div>
    </div>
  );
}