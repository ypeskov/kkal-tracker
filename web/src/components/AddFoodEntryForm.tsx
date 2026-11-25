import { useMemo, useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Ingredient } from '@/api/ingredients';
import FoodAutocomplete from './FoodAutocomplete';
import CalculatorInput from './CalculatorInput';
import { fiberToInput, parseFiberInput } from '@/utils/fiber';

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
  const [fiber, setFiber] = useState('');
  const [fiberInvalid, setFiberInvalid] = useState(false);
  const foodInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    foodInputRef.current?.focus();
  }, []);

  const totalCalories = useMemo(() => {
    const weightNum = parseFloat(weight) || 0;
    const kcalNum = parseFloat(kcalPer100g) || 0;
    return Math.round((weightNum * kcalNum) / 100);
  }, [weight, kcalPer100g]);

  const handleFiberChange = (value: string) => {
    setFiber(value);
    setFiberInvalid(false);
  };

  const handleIngredientSelect = (ingredient: Ingredient) => {
    setKcalPer100g(ingredient.kcalPer100g.toString());
    setFats(ingredient.fats ? ingredient.fats.toString() : '');
    setCarbs(ingredient.carbs ? ingredient.carbs.toString() : '');
    setProteins(ingredient.proteins ? ingredient.proteins.toString() : '');
    // Unlike the macros, a fibre of 0 is a known value and stays in the field
    handleFiberChange(fiberToInput(ingredient.fiber));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!foodName || !weight || !kcalPer100g) return;

    // An empty field is unknown (not sent), 0 is a value; out of range is rejected here, not by the server
    const fiberValue = parseFiberInput(fiber);
    if (fiberValue === null) {
      setFiberInvalid(true);
      return;
    }

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
    if (fiberValue !== undefined) entryData.fiber = fiberValue;

    onSubmit(entryData);

    setFoodName('');
    setWeight('');
    setKcalPer100g('');
    setFats('');
    setCarbs('');
    setProteins('');
    setFiber('');

    foodInputRef.current?.focus();
  };

  const isButtonDisabled = isSubmitting || !foodName || !weight || !kcalPer100g;

  return (
    <section className="mb-8">
      <h2 className="text-xl font-semibold text-gray-800 mb-4">{t('dashboard.addFoodEntry')}</h2>
      <form onSubmit={handleSubmit} className="bg-gray-50 p-3 rounded-lg border border-gray-200">
        {/* Desktop: all fields in grid layout, Mobile: standard responsive */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
          {/* Food name - full width on desktop and mobile */}
          <div className="flex items-center gap-2 md:col-span-2">
            <label htmlFor="foodName" className="font-medium text-gray-800 text-sm w-1/5">{t('dashboard.foodName')}:</label>
            <div className="w-4/5">
              <FoodAutocomplete
                ref={foodInputRef}
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
              <div className="flex justify-between gap-2">
                <div className="w-[45%]">
                  <CalculatorInput
                    id="weight"
                    value={weight}
                    onChange={setWeight}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    required
                  />
                </div>
                <div className="w-[45%]">
                  <CalculatorInput
                    id="kcalPer100g"
                    value={kcalPer100g}
                    onChange={setKcalPer100g}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    required
                  />
                </div>
              </div>
            </div>

            {/* Desktop weight field */}
            <div className="hidden lg:flex lg:flex-col">
              <label htmlFor="weight-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.weight')}:</label>
              <CalculatorInput
                id="weight-desktop"
                value={weight}
                onChange={setWeight}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
              />
            </div>
          </div>

          {/* Calories - desktop only */}
          <div className="hidden lg:flex lg:flex-col">
            <label htmlFor="kcalPer100g-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.kcalPer100g')}:</label>
            <CalculatorInput
              id="kcalPer100g-desktop"
              value={kcalPer100g}
              onChange={setKcalPer100g}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              required
            />
          </div>
        </div>

        {/* Mobile: Fats+Carbs - labels on top, inputs below */}
        <div className="lg:hidden mb-4">
          <div className="flex justify-between mb-1">
            <label htmlFor="fats" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.fats')}:</label>
            <label htmlFor="carbs" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.carbs')}:</label>
          </div>
          <div className="flex justify-between gap-2">
            <div className="w-[45%]">
              <CalculatorInput
                id="fats"
                value={fats}
                onChange={setFats}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div className="w-[45%]">
              <CalculatorInput
                id="carbs"
                value={carbs}
                onChange={setCarbs}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>
        </div>

        {/* Mobile: Proteins+Fibre - labels on top, inputs below */}
        <div className="lg:hidden mb-4">
          <div className="flex justify-between mb-1">
            <label htmlFor="proteins" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.proteins')}:</label>
            <label htmlFor="fiber" className="font-medium text-gray-800 text-sm w-[45%]">{t('dashboard.fiber')}:</label>
          </div>
          <div className="flex justify-between gap-2">
            <div className="w-[45%]">
              <CalculatorInput
                id="proteins"
                value={proteins}
                onChange={setProteins}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div className="w-[45%]">
              <CalculatorInput
                id="fiber"
                value={fiber}
                onChange={handleFiberChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
          </div>
        </div>

        {/* Desktop: all 4 in one row */}
        <div className="hidden lg:grid lg:grid-cols-4 gap-4 mb-4">
          <div className="flex flex-col">
            <label htmlFor="fats-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.fats')}:</label>
            <CalculatorInput
              id="fats-desktop"
              value={fats}
              onChange={setFats}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <div className="flex flex-col">
            <label htmlFor="carbs-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.carbs')}:</label>
            <CalculatorInput
              id="carbs-desktop"
              value={carbs}
              onChange={setCarbs}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <div className="flex flex-col">
            <label htmlFor="proteins-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.proteins')}:</label>
            <CalculatorInput
              id="proteins-desktop"
              value={proteins}
              onChange={setProteins}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          <div className="flex flex-col">
            <label htmlFor="fiber-desktop" className="font-medium mb-1 text-gray-800 text-sm">{t('dashboard.fiber')}:</label>
            <CalculatorInput
              id="fiber-desktop"
              value={fiber}
              onChange={handleFiberChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {fiberInvalid && (
          <p role="alert" className="mb-4 text-sm text-red-600">{t('common.fiberInvalid')}</p>
        )}

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