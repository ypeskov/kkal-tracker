import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Ingredient, UpdateIngredientData } from '../api/ingredients';
// Modal.css imports removed - using Tailwind CSS

interface EditIngredientModalProps {
  ingredient: Ingredient;
  onUpdate: (id: number, data: UpdateIngredientData) => void;
  onDelete: (id: number) => void;
  onCancel: () => void;
  isUpdating: boolean;
  isDeleting: boolean;
}

export default function EditIngredientModal({
  ingredient,
  onUpdate,
  onDelete,
  onCancel,
  isUpdating,
  isDeleting
}: EditIngredientModalProps) {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [kcalPer100g, setKcalPer100g] = useState('');
  const [proteins, setProteins] = useState('');
  const [carbs, setCarbs] = useState('');
  const [fats, setFats] = useState('');

  const handleNumericInput = (value: string, setter: (value: string) => void) => {
    // Replace all commas with dots
    const normalizedValue = value.replace(/,/g, '.');
    // Only allow valid number format
    if (/^\d*\.?\d*$/.test(normalizedValue) || normalizedValue === '') {
      setter(normalizedValue);
    }
  };

  useEffect(() => {
    if (ingredient) {
      setName(ingredient.name);
      setKcalPer100g(ingredient.kcalPer100g.toString());
      setProteins(ingredient.proteins?.toString() || '');
      setCarbs(ingredient.carbs?.toString() || '');
      setFats(ingredient.fats?.toString() || '');
    }
  }, [ingredient]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const data: UpdateIngredientData = {
      name: name.trim(),
      kcalPer100g: parseFloat(kcalPer100g),
      proteins: proteins ? parseFloat(proteins) : undefined,
      carbs: carbs ? parseFloat(carbs) : undefined,
      fats: fats ? parseFloat(fats) : undefined,
    };

    onUpdate(ingredient.id, data);
  };

  const handleDelete = () => {
    onDelete(ingredient.id);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn px-4" onClick={onCancel}>
      <div className="bg-white rounded-lg w-full md:w-[600px] lg:w-[700px] max-h-[90vh] overflow-y-auto shadow-xl animate-slideUp" onClick={(e) => e.stopPropagation()}>
        <div className="flex justify-between items-center p-5 border-b border-gray-200">
          <h2 className="text-xl font-medium text-gray-800 m-0">{t('foodList.editIngredient')}</h2>
          <button
            className="bg-transparent border-none text-2xl text-gray-400 cursor-pointer p-0 w-7 h-7 flex items-center justify-center transition-colors duration-200 hover:text-gray-600"
            onClick={onCancel}
          >
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5">
          <div className="mb-4">
            <label htmlFor="name" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.name')}</label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
            />
          </div>

          <div className="mb-4">
            <label htmlFor="kcalPer100g" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.caloriesPer100g')}</label>
            <input
              id="kcalPer100g"
              type="text"
              inputMode="decimal"
              value={kcalPer100g}
              onChange={(e) => handleNumericInput(e.target.value, setKcalPer100g)}
                            required
              className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
            />
          </div>

          {/* Mobile: Proteins+Carbs in 2 cols, Fats separate | Desktop: all 3 in one row */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div className="mb-4">
              <label htmlFor="proteins" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.proteins')}</label>
              <input
                id="proteins"
                type="text"
              inputMode="decimal"
                    value={proteins}
                onChange={(e) => handleNumericInput(e.target.value, setProteins)}
                                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
              />
            </div>

            <div className="mb-4">
              <label htmlFor="carbs" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.carbs')}</label>
              <input
                id="carbs"
                type="text"
              inputMode="decimal"
                    value={carbs}
                onChange={(e) => handleNumericInput(e.target.value, setCarbs)}
                                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
              />
            </div>

            {/* On mobile spans 1 col (50% width), on desktop takes 1 col */}
            <div className="mb-4 col-span-1 md:col-span-1">
              <label htmlFor="fats" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.fats')}</label>
              <input
                id="fats"
                type="text"
              inputMode="decimal"
                    value={fats}
                onChange={(e) => handleNumericInput(e.target.value, setFats)}
                                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
              />
            </div>
          </div>

          <div className="flex flex-col sm:flex-row sm:justify-between items-center mt-8 pt-5 border-t border-gray-200 gap-3">
            {/* Update and Cancel buttons together */}
            <div className="flex gap-2 w-full sm:w-auto">
              <button
                type="submit"
                className="btn-primary px-3 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed flex-1 sm:flex-initial"
                disabled={isUpdating || isDeleting}
              >
                {isUpdating ? t('common.updating') : t('common.update')}
              </button>
              <button
                type="button"
                className="btn-secondary px-3 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed flex-1 sm:flex-initial"
                onClick={onCancel}
                disabled={isUpdating || isDeleting}
              >
                {t('common.cancel')}
              </button>
            </div>

            {/* Delete button */}
            <button
              type="button"
              className="btn-danger px-3 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed w-full sm:w-auto"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? t('common.deleting') : t('common.delete')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}