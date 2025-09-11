import { useMemo, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

interface EditEntryModalProps {
  entry: any;
  onUpdate: (id: number, entry: any) => void;
  onCancel: () => void;
  onDelete: () => void;
  isUpdating: boolean;
  isDeleting: boolean;
}

export default function EditEntryModal({ entry, onUpdate, onCancel, onDelete, isUpdating, isDeleting }: EditEntryModalProps) {
  const { t } = useTranslation();
  const [editFoodName, setEditFoodName] = useState('');
  const [editWeight, setEditWeight] = useState('');
  const [editKcalPer100g, setEditKcalPer100g] = useState('');
  const [editFats, setEditFats] = useState('');
  const [editCarbs, setEditCarbs] = useState('');
  const [editProteins, setEditProteins] = useState('');
  const [editDate, setEditDate] = useState('');
  const [editTime, setEditTime] = useState('');

  useEffect(() => {
    if (entry) {
      setEditFoodName(entry.food);
      setEditWeight(entry.weight.toString());
      setEditKcalPer100g(entry.kcalPer100g.toString());
      setEditFats(entry.fats ? entry.fats.toString() : '');
      setEditCarbs(entry.carbs ? entry.carbs.toString() : '');
      setEditProteins(entry.proteins ? entry.proteins.toString() : '');
      const mealDate = new Date(entry.meal_datetime);
      const dateString = mealDate.toISOString().split('T')[0];
      const timeString = mealDate.toTimeString().split(' ')[0].substring(0, 5);
      setEditDate(dateString);
      setEditTime(timeString);
    }
  }, [entry]);

  const editTotalCalories = useMemo(() => {
    const weightNum = parseFloat(editWeight) || 0;
    const kcalNum = parseFloat(editKcalPer100g) || 0;
    return Math.round((weightNum * kcalNum) / 100);
  }, [editWeight, editKcalPer100g]);

  const handleUpdateSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!editFoodName || !editWeight || !editKcalPer100g || !entry || !editDate || !editTime) return;

    const mealDateTime = new Date(`${editDate}T${editTime}:00`).toISOString();

    const entryData: any = {
      food: editFoodName,
      weight: parseFloat(editWeight),
      kcalPer100g: parseFloat(editKcalPer100g),
      calories: editTotalCalories,
      meal_datetime: mealDateTime,
    };

    if (editFats && parseFloat(editFats) > 0) entryData.fats = parseFloat(editFats);
    if (editCarbs && parseFloat(editCarbs) > 0) entryData.carbs = parseFloat(editCarbs);
    if (editProteins && parseFloat(editProteins) > 0) entryData.proteins = parseFloat(editProteins);

    onUpdate(entry.id, entryData);
  };

  if (!entry) return null;

  return (
    <div style={{ 
      position: 'fixed', 
      top: 0, 
      left: 0, 
      right: 0, 
      bottom: 0, 
      backgroundColor: 'rgba(0,0,0,0.5)', 
      zIndex: 1000,
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      padding: '1rem'
    }}>
      <div style={{ 
        backgroundColor: 'white', 
        borderRadius: '8px', 
        padding: '2rem', 
        maxWidth: '500px', 
        width: '100%',
        maxHeight: '90vh',
        overflowY: 'auto'
      }}>
        <h2 style={{ marginTop: 0 }}>{t('dashboard.editEntry')}</h2>
        
        <form onSubmit={handleUpdateSubmit}>
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="editFoodName">{t('dashboard.foodName')}:</label>
              <input
                type="text"
                id="editFoodName"
                value={editFoodName}
                onChange={(e) => setEditFoodName(e.target.value)}
                required
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="editDate">{t('dashboard.date')}:</label>
              <input
                type="date"
                id="editDate"
                value={editDate}
                onChange={(e) => setEditDate(e.target.value)}
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="editTime">{t('dashboard.time')}:</label>
              <input
                type="time"
                id="editTime"
                value={editTime}
                onChange={(e) => setEditTime(e.target.value)}
                required
              />
            </div>
          </div>
          
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="editWeight">{t('dashboard.weight')}:</label>
              <input
                type="number"
                id="editWeight"
                value={editWeight}
                onChange={(e) => setEditWeight(e.target.value)}
                required
                min="0"
                step="0.1"
              />
            </div>

            <div className="form-group">
              <label htmlFor="editKcalPer100g">{t('dashboard.kcalPer100g')}:</label>
              <input
                type="number"
                id="editKcalPer100g"
                value={editKcalPer100g}
                onChange={(e) => setEditKcalPer100g(e.target.value)}
                required
                min="0"
                step="0.1"
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="editFats">{t('dashboard.fats')}:</label>
              <input
                type="number"
                id="editFats"
                value={editFats}
                onChange={(e) => setEditFats(e.target.value)}
                min="0"
                step="0.1"
              />
            </div>

            <div className="form-group">
              <label htmlFor="editCarbs">{t('dashboard.carbs')}:</label>
              <input
                type="number"
                id="editCarbs"
                value={editCarbs}
                onChange={(e) => setEditCarbs(e.target.value)}
                min="0"
                step="0.1"
              />
            </div>

            <div className="form-group">
              <label htmlFor="editProteins">{t('dashboard.proteins')}:</label>
              <input
                type="number"
                id="editProteins"
                value={editProteins}
                onChange={(e) => setEditProteins(e.target.value)}
                min="0"
                step="0.1"
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="editTotalCalories">{t('dashboard.totalCalories')}:</label>
              <input
                type="number"
                id="editTotalCalories"
                value={editTotalCalories}
                readOnly
                className="readonly-field"
              />
            </div>
          </div>

          <div style={{ 
            marginTop: '2rem', 
            display: 'flex', 
            justifyContent: 'space-between', 
            alignItems: 'center',
            width: '100%',
            gap: '1rem'
          }} className="modal-button-container">
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button 
                type="submit" 
                className="modal-btn"
                disabled={isUpdating || isDeleting || !editFoodName || !editWeight || !editKcalPer100g}
                style={{ backgroundColor: '#28a745' }}
              >
                {isUpdating ? t('common.loading') + '...' : t('dashboard.updateEntry')}
              </button>
              <button 
                type="button" 
                onClick={onCancel}
                className="modal-btn"
                style={{ backgroundColor: '#6c757d' }}
                disabled={isUpdating || isDeleting}
              >
                {t('dashboard.cancel')}
              </button>
            </div>
            
            <button 
              type="button" 
              onClick={onDelete}
              className="modal-btn"
              style={{ backgroundColor: '#dc3545' }}
              disabled={isUpdating || isDeleting}
            >
              {t('dashboard.delete')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}