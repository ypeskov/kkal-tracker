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
    <section style={{ marginBottom: '2rem' }}>
      <h2>{t('dashboard.addFoodEntry')}</h2>
      <form onSubmit={handleSubmit} className="food-form">
        <div className="form-row">
          <div className="form-group">
            <label htmlFor="foodName">{t('dashboard.foodName')}:</label>
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

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="weight">{t('dashboard.weight')}:</label>
            <input
              type="number"
              id="weight"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
              required
              min="0"
              step="0.1"
            />
          </div>

          <div className="form-group">
            <label htmlFor="kcalPer100g">{t('dashboard.kcalPer100g')}:</label>
            <input
              type="number"
              id="kcalPer100g"
              value={kcalPer100g}
              onChange={(e) => setKcalPer100g(e.target.value)}
              required
              min="0"
              step="0.1"
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="fats">{t('dashboard.fats')}:</label>
            <input
              type="number"
              id="fats"
              value={fats}
              onChange={(e) => setFats(e.target.value)}
              min="0"
              step="0.1"
            />
          </div>

          <div className="form-group">
            <label htmlFor="carbs">{t('dashboard.carbs')}:</label>
            <input
              type="number"
              id="carbs"
              value={carbs}
              onChange={(e) => setCarbs(e.target.value)}
              min="0"
              step="0.1"
            />
          </div>

          <div className="form-group">
            <label htmlFor="proteins">{t('dashboard.proteins')}:</label>
            <input
              type="number"
              id="proteins"
              value={proteins}
              onChange={(e) => setProteins(e.target.value)}
              min="0"
              step="0.1"
            />
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="totalCalories">{t('dashboard.totalCalories')}:</label>
            <input
              type="number"
              id="totalCalories"
              value={totalCalories}
              readOnly
              className="readonly-field"
            />
          </div>

          <button
            type="submit"
            className="btn submit-btn"
            disabled={isButtonDisabled}
          >
            {isSubmitting ? t('common.loading') + '...' : t('dashboard.addEntry')}
          </button>
        </div>
      </form>
    </section>
  );
}