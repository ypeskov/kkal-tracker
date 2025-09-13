import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CreateIngredientData } from '../api/ingredients';
// Modal.css imports removed - using Tailwind CSS

interface AddIngredientModalProps {
  onCreate: (data: CreateIngredientData) => void;
  onCancel: () => void;
  isCreating: boolean;
}

export default function AddIngredientModal({
  onCreate,
  onCancel,
  isCreating
}: AddIngredientModalProps) {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [kcalPer100g, setKcalPer100g] = useState('');
  const [proteins, setProteins] = useState('');
  const [carbs, setCarbs] = useState('');
  const [fats, setFats] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    const data: CreateIngredientData = {
      name: name.trim(),
      kcalPer100g: parseFloat(kcalPer100g),
      proteins: proteins ? parseFloat(proteins) : undefined,
      carbs: carbs ? parseFloat(carbs) : undefined,
      fats: fats ? parseFloat(fats) : undefined,
    };

    onCreate(data);
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 animate-fadeIn" onClick={onCancel}>
      <div className="bg-white rounded-lg max-w-2xl w-full max-w-[90%] max-h-[90vh] overflow-y-auto shadow-xl animate-slideUp" onClick={(e) => e.stopPropagation()}>
        <div className="flex justify-between items-center p-5 border-b border-gray-200">
          <h2 className="text-xl font-medium text-gray-800 m-0">{t('foodList.addNewIngredient')}</h2>
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
              autoFocus
            />
          </div>

          <div className="mb-4">
            <label htmlFor="kcalPer100g" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.caloriesPer100g')}</label>
            <input
              id="kcalPer100g"
              type="number"
              step="0.1"
              min="0"
              value={kcalPer100g}
              onChange={(e) => setKcalPer100g(e.target.value)}
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
                type="number"
                step="0.1"
                min="0"
                value={proteins}
                onChange={(e) => setProteins(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                placeholder={t('foodList.optional')}
              />
            </div>

            <div className="mb-4">
              <label htmlFor="carbs" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.carbs')}</label>
              <input
                id="carbs"
                type="number"
                step="0.1"
                min="0"
                value={carbs}
                onChange={(e) => setCarbs(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                placeholder={t('foodList.optional')}
              />
            </div>

            {/* On mobile spans 1 col (50% width), on desktop takes 1 col */}
            <div className="mb-4 col-span-1 md:col-span-1">
              <label htmlFor="fats" className="block mb-1 text-gray-600 text-sm font-medium">{t('foodList.fats')}</label>
              <input
                id="fats"
                type="number"
                step="0.1"
                min="0"
                value={fats}
                onChange={(e) => setFats(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-green-500 focus:ring-2 focus:ring-green-500 focus:ring-opacity-10 text-sm transition-colors placeholder-gray-400"
                placeholder={t('foodList.optional')}
              />
            </div>
          </div>

          <div className="flex justify-center items-center mt-8 pt-5 border-t border-gray-200">
            <div className="flex gap-3">
              <button
                type="submit"
                className="btn-primary px-4 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed"
                disabled={isCreating}
              >
                {isCreating ? t('common.creating') : t('common.create')}
              </button>
              <button
                type="button"
                className="btn-secondary px-4 py-2 text-sm font-medium disabled:opacity-60 disabled:cursor-not-allowed"
                onClick={onCancel}
                disabled={isCreating}
              >
                {t('common.cancel')}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}