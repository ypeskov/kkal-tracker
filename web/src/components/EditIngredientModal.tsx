import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Ingredient, UpdateIngredientData } from '../api/ingredients';
import './Modal.css';

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
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>{t('foodList.editIngredient')}</h2>
          <button className="modal-close" onClick={onCancel}>×</button>
        </div>

        <form onSubmit={handleSubmit} className="modal-form">
          <div className="form-group">
            <label htmlFor="name">{t('foodList.name')}</label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="form-input"
            />
          </div>

          <div className="form-group">
            <label htmlFor="kcalPer100g">{t('foodList.caloriesPer100g')}</label>
            <input
              id="kcalPer100g"
              type="number"
              step="0.1"
              min="0"
              value={kcalPer100g}
              onChange={(e) => setKcalPer100g(e.target.value)}
              required
              className="form-input"
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="proteins">{t('foodList.proteins')}</label>
              <input
                id="proteins"
                type="number"
                step="0.1"
                min="0"
                value={proteins}
                onChange={(e) => setProteins(e.target.value)}
                className="form-input"
              />
            </div>

            <div className="form-group">
              <label htmlFor="carbs">{t('foodList.carbs')}</label>
              <input
                id="carbs"
                type="number"
                step="0.1"
                min="0"
                value={carbs}
                onChange={(e) => setCarbs(e.target.value)}
                className="form-input"
              />
            </div>

            <div className="form-group">
              <label htmlFor="fats">{t('foodList.fats')}</label>
              <input
                id="fats"
                type="number"
                step="0.1"
                min="0"
                value={fats}
                onChange={(e) => setFats(e.target.value)}
                className="form-input"
              />
            </div>
          </div>

          <div className="modal-actions">
            <button
              type="button"
              className="btn btn-danger"
              onClick={handleDelete}
              disabled={isDeleting}
            >
              {isDeleting ? t('common.deleting') : t('common.delete')}
            </button>
            
            <div className="modal-actions-right">
              <button
                type="button"
                className="btn btn-secondary"
                onClick={onCancel}
                disabled={isUpdating || isDeleting}
              >
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                className="btn btn-primary"
                disabled={isUpdating || isDeleting}
              >
                {isUpdating ? t('common.updating') : t('common.update')}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}