import { useMemo, useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { handleNumericInput } from '@/utils/numericInput';

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
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn px-4">
      <div className="bg-white rounded-lg w-full md:w-[600px] lg:w-[700px] max-h-[90vh] overflow-y-auto shadow-xl animate-slideUp">
        <div className="flex justify-between items-center p-5 border-b border-gray-200">
          <h2 className="text-xl font-medium text-gray-800 m-0">{t('dashboard.editEntry')}</h2>
        </div>
        
        <form onSubmit={handleUpdateSubmit} className="p-5">
          <div className="flex flex-col md:flex-row gap-4 mb-4 items-stretch md:items-end">
            <div className="flex flex-col flex-1">
              <label htmlFor="editFoodName" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.foodName')}:</label>
              <input
                type="text"
                id="editFoodName"
                value={editFoodName}
                onChange={(e) => setEditFoodName(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                required
              />
            </div>
          </div>

          <div className="flex flex-col sm:flex-row gap-4 mb-4">
            <div className="flex flex-col flex-1">
              <label htmlFor="editDate" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.date')}:</label>
              <input
                type="date"
                id="editDate"
                value={editDate}
                onChange={(e) => setEditDate(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                required
              />
            </div>

            <div className="flex flex-col flex-1">
              <label htmlFor="editTime" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.time')}:</label>
              <input
                type="time"
                id="editTime"
                value={editTime}
                onChange={(e) => setEditTime(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                required
              />
            </div>
          </div>
          
          <div className="flex flex-col sm:flex-row gap-4 mb-4">
            <div className="flex flex-col flex-1">
              <label htmlFor="editWeight" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.weight')}:</label>
              <input
                type="text"
                inputMode="decimal"
                id="editWeight"
                value={editWeight}
                onChange={(e) => handleNumericInput(e.target.value, setEditWeight)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                required
              />
            </div>

            <div className="flex flex-col flex-1">
              <label htmlFor="editKcalPer100g" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.kcalPer100g')}:</label>
              <input
                type="text"
                inputMode="decimal"
                id="editKcalPer100g"
                value={editKcalPer100g}
                onChange={(e) => handleNumericInput(e.target.value, setEditKcalPer100g)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                required
              />
            </div>
          </div>

          {/* Mobile: Fats+Carbs in 2 cols, Proteins separate | Desktop: all 3 in one row */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
            <div className="flex flex-col">
              <label htmlFor="editFats" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.fats')}:</label>
              <input
                type="text"
                inputMode="decimal"
                id="editFats"
                value={editFats}
                onChange={(e) => handleNumericInput(e.target.value, setEditFats)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
              />
            </div>

            <div className="flex flex-col">
              <label htmlFor="editCarbs" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.carbs')}:</label>
              <input
                type="text"
                inputMode="decimal"
                id="editCarbs"
                value={editCarbs}
                onChange={(e) => handleNumericInput(e.target.value, setEditCarbs)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
              />
            </div>

            {/* On mobile spans 1 col (50% width), on desktop takes 1 col */}
            <div className="flex flex-col col-span-1 md:col-span-1">
              <label htmlFor="editProteins" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.proteins')}:</label>
              <input
                type="text"
                inputMode="decimal"
                id="editProteins"
                value={editProteins}
                onChange={(e) => handleNumericInput(e.target.value, setEditProteins)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
              />
            </div>
          </div>

          <div className="flex flex-col md:flex-row gap-4 mb-4 items-stretch md:items-end">
            <div className="flex flex-col flex-1">
              <label htmlFor="editTotalCalories" className="block mb-1 text-gray-600 text-sm font-medium">{t('dashboard.totalCalories')}:</label>
              <input
                type="number"
                id="editTotalCalories"
                value={editTotalCalories}
                readOnly
                className="w-full px-3 py-2 border border-gray-300 rounded bg-gray-100 cursor-not-allowed text-gray-600"
              />
            </div>
          </div>

          <div className="flex flex-col sm:flex-row sm:justify-between items-center mt-8 pt-5 border-t border-gray-200 gap-3">
            {/* Update and Cancel buttons together */}
            <div className="flex gap-2 w-full sm:w-auto">
              <button
                type="submit"
                className="btn-primary px-3 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed flex-1 sm:flex-initial"
                disabled={isUpdating || isDeleting || !editFoodName || !editWeight || !editKcalPer100g}
              >
                {isUpdating ? t('common.loading') + '...' : t('dashboard.updateEntry')}
              </button>
              <button
                type="button"
                onClick={onCancel}
                className="btn-secondary px-3 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed flex-1 sm:flex-initial"
                disabled={isUpdating || isDeleting}
              >
                {t('dashboard.cancel')}
              </button>
            </div>

            {/* Delete button */}
            <button
              type="button"
              onClick={onDelete}
              className="btn-danger px-3 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed w-full sm:w-auto"
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