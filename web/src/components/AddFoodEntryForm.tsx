import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Ingredient } from '../api/ingredients';
import FoodAutocomplete from './FoodAutocomplete';

interface AddFoodEntryFormProps {
  onSubmit: (entry: any) => void;
  isSubmitting: boolean;
}

export default function AddFoodEntryForm({ onSubmit, isSubmitting }: AddFoodEntryFormProps) {
  const { t } = useTranslation();
  const [foodName, setFoodName] = useState('');
  const [weight, setWeight] = useState('');
  const [kcalPer100g, setKcalPer100g] = useState('');
  const [fats, setFats] = useState('');
  const [carbs, setCarbs] = useState('');
  const [proteins, setProteins] = useState('');

  const totalCalories = useMemo(() => {
    const weightNum = parseFloat(weight) || 0;
    const kcalNum = parseFloat(kcalPer100g) || 0;
    return Math.round((weightNum * kcalNum) / 100);
  }, [weight, kcalPer100g]);

  const handleIngredientSelect = (ingredient: Ingredient) => {
    setKcalPer100g(ingredient.kcalPer100g.toString());
    setFats(ingredient.fats ? ingredient.fats.toString() : '');
    setCarbs(ingredient.carbs ? ingredient.carbs.toString() : '');
    setProteins(ingredient.proteins ? ingredient.proteins.toString() : '');
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!foodName || !weight || !kcalPer100g) return;

    const entryData: any = {
      food: foodName,
      weight: parseFloat(weight),
      kcalPer100g: parseFloat(kcalPer100g),
      calories: totalCalories,
      meal_datetime: new Date().toISOString(),
    };

    if (fats && parseFloat(fats) > 0) entryData.fats = parseFloat(fats);
    if (carbs && parseFloat(carbs) > 0) entryData.carbs = parseFloat(carbs);
    if (proteins && parseFloat(proteins) > 0) entryData.proteins = parseFloat(proteins);

    onSubmit(entryData);

    setFoodName('');
    setWeight('');
    setKcalPer100g('');
    setFats('');
    setCarbs('');
    setProteins('');
  };

  const isButtonDisabled = isSubmitting || !foodName || !weight || !kcalPer100g;

  return (
    <section className="mb-8">
      <h2 className="text-xl font-semibold text-gray-800 mb-4">{t('dashboard.addFoodEntry')}</h2>
      <form onSubmit={handleSubmit} className="bg-gray-50 p-6 rounded-lg border border-gray-200">
        <div className="flex flex-col gap-4 mb-4 md:flex-row md:items-end">
          <div className="flex flex-col flex-1">
            <label htmlFor="foodName" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.foodName')}:</label>
            <FoodAutocomplete
              id="foodName"
              value={foodName}
              onChange={setFoodName}
              onSelect={handleIngredientSelect}
              placeholder={t('dashboard.foodName')}
              required
            />
          </div>
        </div>

        <div className="flex flex-col gap-4 mb-4 sm:flex-row">
          <div className="flex flex-col flex-1">
            <label htmlFor="weight" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.weight')}:</label>
            <input
              type="number"
              id="weight"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
              min="0"
              step="0.1"
            />
          </div>

          <div className="flex flex-col flex-1">
            <label htmlFor="kcalPer100g" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.kcalPer100g')}:</label>
            <input
              type="number"
              id="kcalPer100g"
              value={kcalPer100g}
              onChange={(e) => setKcalPer100g(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
              min="0"
              step="0.1"
            />
          </div>
        </div>

        {/* Жиры и углеводы в одной строке */}
        <div className="flex flex-col gap-4 mb-4 sm:flex-row">
          <div className="flex flex-col flex-1">
            <label htmlFor="fats" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.fats')}:</label>
            <input
              type="number"
              id="fats"
              value={fats}
              onChange={(e) => setFats(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>

          <div className="flex flex-col flex-1">
            <label htmlFor="carbs" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.carbs')}:</label>
            <input
              type="number"
              id="carbs"
              value={carbs}
              onChange={(e) => setCarbs(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>
        </div>

        {/* Белки в отдельной строке, 50% ширины */}
        <div className="mb-4">
          <div className="flex flex-col w-full sm:w-1/2">
            <label htmlFor="proteins" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.proteins')}:</label>
            <input
              type="number"
              id="proteins"
              value={proteins}
              onChange={(e) => setProteins(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>
        </div>

        <div className="flex flex-col gap-4 sm:flex-row sm:items-end">
          <div className="flex flex-col flex-1">
            <label htmlFor="totalCalories" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.totalCalories')}:</label>
            <input
              type="number"
              id="totalCalories"
              value={totalCalories}
              readOnly
              className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100 cursor-not-allowed text-gray-600"
            />
          </div>

          <button
            type="submit"
            className="btn-primary disabled:opacity-60 disabled:cursor-not-allowed h-10 px-6 min-w-24"
            disabled={isButtonDisabled}
          >
            {isSubmitting ? t('common.loading') + '...' : t('dashboard.addEntry')}
          </button>
        </div>
      </form>
    </section>
  );
}