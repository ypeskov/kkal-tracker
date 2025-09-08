import { useTranslation } from 'react-i18next';

interface DeleteConfirmationDialogProps {
  entry: any;
  onConfirm: () => void;
  onCancel: () => void;
  isDeleting: boolean;
}

export default function DeleteConfirmationDialog({ entry, onConfirm, onCancel, isDeleting }: DeleteConfirmationDialogProps) {
  const { t } = useTranslation();

  if (!entry) return null;

  return (
    <div style={{ 
      position: 'fixed', 
      top: 0, 
      left: 0, 
      right: 0, 
      bottom: 0, 
      backgroundColor: 'rgba(0,0,0,0.5)', 
      zIndex: 1001,
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      padding: '1rem'
    }}>
      <div style={{ 
        backgroundColor: 'white', 
        borderRadius: '8px', 
        padding: '2rem', 
        maxWidth: '400px', 
        width: '100%',
        textAlign: 'center'
      }}>
        <h3 style={{ marginTop: 0, color: '#dc3545' }}>{t('dashboard.confirmDelete')}</h3>
        <p style={{ marginBottom: '2rem' }}>
          {t('dashboard.deleteWarning', { foodName: entry?.food })}
        </p>
        
        <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem' }}>
          <button 
            type="button" 
            onClick={onCancel}
            className="btn"
            style={{ backgroundColor: '#6c757d' }}
            disabled={isDeleting}
          >
            {t('dashboard.cancel')}
          </button>
          <button 
            type="button" 
            onClick={onConfirm}
            className="btn"
            style={{ backgroundColor: '#dc3545' }}
            disabled={isDeleting}
          >
            {isDeleting ? t('common.loading') + '...' : t('dashboard.confirmDeleteButton')}
          </button>
        </div>
      </div>
    </div>
  );
}