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
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h2>{t('dashboard.editEntry')}</h2>
        </div>
        
        <form onSubmit={handleUpdateSubmit} className="modal-form">
          <div className="form-row modal-food-name-row">
            <div className="form-group">
              <label htmlFor="editFoodName">{t('dashboard.foodName')}:</label>
              <input
                type="text"
                id="editFoodName"
                value={editFoodName}
                onChange={(e) => setEditFoodName(e.target.value)}
                className="form-input"
                required
              />
            </div>
          </div>

          <div className="form-row modal-date-time-row">
            <div className="form-group">
              <label htmlFor="editDate">{t('dashboard.date')}:</label>
              <input
                type="date"
                id="editDate"
                value={editDate}
                onChange={(e) => setEditDate(e.target.value)}
                className="form-input"
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
                className="form-input"
                required
              />
            </div>
          </div>
          
          <div className="form-row modal-weight-calories-row">
            <div className="form-group">
              <label htmlFor="editWeight">{t('dashboard.weight')}:</label>
              <input
                type="number"
                id="editWeight"
                value={editWeight}
                onChange={(e) => setEditWeight(e.target.value)}
                className="form-input"
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
                className="form-input"
                required
                min="0"
                step="0.1"
              />
            </div>
          </div>

          <div className="form-row modal-fats-carbs-proteins-row">
            <div className="form-group">
              <label htmlFor="editFats">{t('dashboard.fats')}:</label>
              <input
                type="number"
                id="editFats"
                value={editFats}
                onChange={(e) => setEditFats(e.target.value)}
                className="form-input"
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
                className="form-input"
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
                className="form-input"
                min="0"
                step="0.1"
              />
            </div>
          </div>

          <div className="form-row modal-total-calories-row">
            <div className="form-group">
              <label htmlFor="editTotalCalories">{t('dashboard.totalCalories')}:</label>
              <input
                type="number"
                id="editTotalCalories"
                value={editTotalCalories}
                readOnly
                className="form-input form-input--readonly"
              />
            </div>
          </div>

          <div className="modal-actions">
            <div className="modal-actions-right">
              <button
                type="submit"
                className="btn btn--success"
                disabled={isUpdating || isDeleting || !editFoodName || !editWeight || !editKcalPer100g}
              >
                {isUpdating ? t('common.loading') + '...' : t('dashboard.updateEntry')}
              </button>
              <button
                type="button"
                onClick={onCancel}
                className="btn btn--secondary"
                disabled={isUpdating || isDeleting}
              >
                {t('dashboard.cancel')}
              </button>
            </div>
            
            <button
              type="button"
              onClick={onDelete}
              className="btn btn--danger"
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