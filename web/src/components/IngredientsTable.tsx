import { useTranslation } from 'react-i18next';
import { Ingredient } from '@/api/ingredients';

interface IngredientsTableProps {
  ingredients: Ingredient[];
  onRowClick: (ingredient: Ingredient) => void;
}

export default function IngredientsTable({ ingredients, onRowClick }: IngredientsTableProps) {
  const { t } = useTranslation();

  return (
    <div className="bg-white rounded-lg shadow-lg overflow-hidden">
      <table className="w-full border-collapse">
        <thead className="bg-gray-100">
          <tr>
            <th className="p-4 md:p-4 text-left font-semibold text-gray-600 md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">
              {t('foodList.name')}
            </th>
            <th className="p-4 md:p-4 text-left font-semibold text-gray-600 md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">
              {t('foodList.calories')}
            </th>
            <th className="p-4 md:p-4 text-left font-semibold text-gray-600 md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">
              {t('foodList.proteins')}
            </th>
            <th className="p-4 md:p-4 text-left font-semibold text-gray-600 md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">
              {t('foodList.carbs')}
            </th>
            <th className="p-4 md:p-4 text-left font-semibold text-gray-600 md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">
              {t('foodList.fats')}
            </th>
          </tr>
        </thead>
        <tbody>
          {ingredients.length === 0 ? (
            <tr>
              <td colSpan={5} className="text-center text-gray-500 italic py-10 text-base">
                {t('foodList.noIngredients')}
              </td>
            </tr>
          ) : (
            ingredients.map((ingredient) => (
              <tr
                key={ingredient.id}
                className="border-b border-gray-200 cursor-pointer transition-all duration-200 ease-in-out hover:bg-gray-50 md:hover:translate-x-0.5 md:hover:shadow-[inset_3px_0_0_theme(colors.green.500)] active:bg-gray-200"
                onClick={() => onRowClick(ingredient)}
              >
                <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs font-medium text-slate-700">
                  {ingredient.name}
                </td>
                <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">
                  {ingredient.kcalPer100g}
                </td>
                <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">
                  {ingredient.proteins ?? '-'}
                </td>
                <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">
                  {ingredient.carbs ?? '-'}
                </td>
                <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">
                  {ingredient.fats ?? '-'}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
