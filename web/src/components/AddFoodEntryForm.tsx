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
        {/* Desktop: all fields in grid layout, Mobile: standard responsive */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          {/* Food name - full width on desktop and mobile */}
          <div className="flex items-center gap-2 md:col-span-2">
            <label htmlFor="foodName" className="font-medium text-gray-800 text-sm w-1/5">{t('dashboard.foodName')}:</label>
            <div className="w-4/5">
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

          {/* Weight - mobile: labels on top, inputs below */}
          <div className="flex flex-col md:flex-col">
            <div className="lg:hidden mb-4">
              <div className="flex justify-between mb-1">
                <label htmlFor="weight" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.weight')}:</label>
                <label htmlFor="kcalPer100g" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.kcalPer100g')}:</label>
              </div>
              <div className="flex justify-between">
                <input
                  type="number"
                  id="weight"
                  value={weight}
                  onChange={(e) => setWeight(e.target.value)}
                  className="w-[45%] px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                  min="0"
                  step="0.1"
                />
                <input
                  type="number"
                  id="kcalPer100g"
                  value={kcalPer100g}
                  onChange={(e) => setKcalPer100g(e.target.value)}
                  className="w-[45%] px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                  min="0"
                  step="0.1"
                />
              </div>
            </div>

            {/* Desktop weight field */}
            <div className="hidden lg:flex lg:flex-col">
              <label htmlFor="weight-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.weight')}:</label>
              <input
                type="number"
                id="weight-desktop"
                value={weight}
                onChange={(e) => setWeight(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
                min="0"
                step="0.1"
              />
            </div>
          </div>

          {/* Calories - desktop only */}
          <div className="hidden lg:flex lg:flex-col">
            <label htmlFor="kcalPer100g-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.kcalPer100g')}:</label>
            <input
              type="number"
              id="kcalPer100g-desktop"
              value={kcalPer100g}
              onChange={(e) => setKcalPer100g(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
              min="0"
              step="0.1"
            />
          </div>
        </div>

        {/* Mobile: Fats+Carbs - labels on top, inputs below */}
        <div className="lg:hidden mb-4">
          <div className="flex justify-between mb-1">
            <label htmlFor="fats" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.fats')}:</label>
            <label htmlFor="carbs" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.carbs')}:</label>
          </div>
          <div className="flex justify-between">
            <input
              type="number"
              id="fats"
              value={fats}
              onChange={(e) => setFats(e.target.value)}
              className="w-[45%] px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
            <input
              type="number"
              id="carbs"
              value={carbs}
              onChange={(e) => setCarbs(e.target.value)}
              className="w-[45%] px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>
        </div>

        {/* Mobile: Proteins - two-row layout like paired fields */}
        <div className="lg:hidden mb-4">
          <div className="flex justify-between mb-1">
            <label htmlFor="proteins" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.proteins')}:</label>
            <div className="w-[45%]"></div>
          </div>
          <div className="flex justify-between">
            <input
              type="number"
              id="proteins"
              value={proteins}
              onChange={(e) => setProteins(e.target.value)}
              className="w-[45%] px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
            <div className="w-[45%]"></div>
          </div>
        </div>

        {/* Desktop: all 3 in one row */}
        <div className="hidden lg:grid lg:grid-cols-3 gap-4 mb-4">
          <div className="flex flex-col">
            <label htmlFor="fats-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.fats')}:</label>
            <input
              type="number"
              id="fats-desktop"
              value={fats}
              onChange={(e) => setFats(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>

          <div className="flex flex-col">
            <label htmlFor="carbs-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.carbs')}:</label>
            <input
              type="number"
              id="carbs-desktop"
              value={carbs}
              onChange={(e) => setCarbs(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>

          {/* On mobile spans 1 col (50% width), on desktop takes 1 col */}
          <div className="flex flex-col col-span-1 md:col-span-1">
            <label htmlFor="proteins-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.proteins')}:</label>
            <input
              type="number"
              id="proteins-desktop"
              value={proteins}
              onChange={(e) => setProteins(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              min="0"
              step="0.1"
            />
          </div>
        </div>

        {/* Mobile: Total calories and submit button */}
        <div className="flex items-center gap-2 mb-4 lg:hidden">
          <label className="font-medium text-gray-800 text-sm w-1/5">{t('dashboard.totalCalories')}:</label>
          <div className="font-bold text-lg text-green-600 flex-1">
            {totalCalories} {t('dashboard.kcal')}
          </div>
          <button
            type="submit"
            className="btn-primary disabled:opacity-60 disabled:cursor-not-allowed h-10 px-6"
            disabled={isButtonDisabled}
          >
            {isSubmitting ? t('common.loading') + '...' : t('dashboard.addEntry')}
          </button>
        </div>

        {/* Desktop: Total calories and submit button */}
        <div className="hidden lg:flex lg:items-center lg:gap-4">
          <div className="flex items-center gap-2 flex-1">
            <label className="font-medium text-gray-800 text-sm w-1/5">{t('dashboard.totalCalories')}:</label>
            <div className="font-bold text-xl text-green-600">
              {totalCalories} {t('dashboard.kcal')}
            </div>
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