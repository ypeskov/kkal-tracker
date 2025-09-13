import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { ingredientService, Ingredient, CreateIngredientData, UpdateIngredientData } from '../api/ingredients';
import EditIngredientModal from '../components/EditIngredientModal';
import AddIngredientModal from '../components/AddIngredientModal';
import DeleteConfirmationDialog from '../components/DeleteConfirmationDialog';
// FoodList.css imports removed - using Tailwind CSS

export default function FoodList() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [selectedIngredient, setSelectedIngredient] = useState<Ingredient | null>(null);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [filterText, setFilterText] = useState('');
  const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);
  const [ingredientToDelete, setIngredientToDelete] = useState<Ingredient | null>(null);

  // Fetch all ingredients
  const { data: ingredients = [], isLoading, error } = useQuery({
    queryKey: ['ingredients'],
    queryFn: ingredientService.loadAndCacheIngredients,
  });

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: CreateIngredientData) => ingredientService.createIngredient(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ingredients'] });
      setIsAddModalOpen(false);
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateIngredientData }) => 
      ingredientService.updateIngredient(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ingredients'] });
      setIsEditModalOpen(false);
      setSelectedIngredient(null);
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => ingredientService.deleteIngredient(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ingredients'] });
      setIsEditModalOpen(false);
      setSelectedIngredient(null);
      setShowDeleteConfirmation(false);
      setIngredientToDelete(null);
    },
  });

  // Filter ingredients based on search text and sort by name
  const filteredIngredients = ingredients
    .filter(ingredient =>
      ingredient.name.toLowerCase().includes(filterText.toLowerCase())
    )
    .sort((a, b) => a.name.localeCompare(b.name));

  const handleRowClick = (ingredient: Ingredient) => {
    setSelectedIngredient(ingredient);
    setIsEditModalOpen(true);
  };

  const handleAddNew = () => {
    setIsAddModalOpen(true);
  };

  const handleCreate = (data: CreateIngredientData) => {
    createMutation.mutate(data);
  };

  const handleUpdate = (id: number, data: UpdateIngredientData) => {
    updateMutation.mutate({ id, data });
  };

  const handleDelete = (id: number) => {
    const ingredient = ingredients.find(i => i.id === id);
    if (ingredient) {
      setIngredientToDelete(ingredient);
      setShowDeleteConfirmation(true);
    }
  };

  const handleConfirmDelete = () => {
    if (ingredientToDelete) {
      deleteMutation.mutate(ingredientToDelete.id);
    }
  };

  const handleCancelDelete = () => {
    setShowDeleteConfirmation(false);
    setIngredientToDelete(null);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <p className="text-gray-600 text-lg">{t('common.loading')}</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-16">
        <p className="text-red-600 text-lg">{t('common.error')}</p>
      </div>
    );
  }

  return (
    <div className="max-w-screen-xl mx-auto p-5 md:p-8">
      <div className="mb-8">
        <h2 className="mb-5 text-gray-800 text-3xl font-semibold">{t('foodList.title')}</h2>

        <div className="flex flex-col md:flex-row gap-5 items-stretch md:items-center mb-5">
          <button
            className="btn-primary px-5 py-2.5 text-sm font-medium uppercase tracking-wide transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg disabled:opacity-60 disabled:cursor-not-allowed"
            onClick={handleAddNew}
          >
            {t('foodList.addNew')}
          </button>
          
          <div className="relative flex-1 max-w-md">
            <input
              type="text"
              placeholder={t('foodList.filterPlaceholder')}
              value={filterText}
              onChange={(e) => setFilterText(e.target.value)}
              className="w-full py-2.5 pl-4 pr-12 border border-gray-300 rounded text-sm transition-colors duration-300 focus:outline-none focus:border-green-500 placeholder-gray-500"
            />
            {filterText && (
              <button
                type="button"
                className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center p-2.5 text-gray-500 text-2xl leading-none transition-colors duration-200 hover:text-gray-700 active:text-gray-800"
                onClick={() => setFilterText('')}
                aria-label={t('common.clear')}
              >
                ×
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-lg overflow-hidden">
        <table className="w-full border-collapse">
          <thead className="bg-gray-100">
            <tr>
              <th className="p-4 md:p-4 p-2.5 text-left font-semibold text-gray-600 text-sm md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">{t('foodList.name')}</th>
              <th className="p-4 md:p-4 p-2.5 text-left font-semibold text-gray-600 text-sm md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">{t('foodList.calories')}</th>
              <th className="p-4 md:p-4 p-2.5 text-left font-semibold text-gray-600 text-sm md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">{t('foodList.proteins')}</th>
              <th className="p-4 md:p-4 p-2.5 text-left font-semibold text-gray-600 text-sm md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">{t('foodList.carbs')}</th>
              <th className="p-4 md:p-4 p-2.5 text-left font-semibold text-gray-600 text-sm md:text-sm text-xs uppercase tracking-wide border-b-2 border-gray-300">{t('foodList.fats')}</th>
            </tr>
          </thead>
          <tbody>
            {filteredIngredients.length === 0 ? (
              <tr>
                <td colSpan={5} className="text-center text-gray-500 italic py-10 text-base">
                  {t('foodList.noIngredients')}
                </td>
              </tr>
            ) : (
              filteredIngredients.map((ingredient) => (
                <tr
                  key={ingredient.id}
                  className="border-b border-gray-200 cursor-pointer transition-all duration-200 ease-in-out hover:bg-gray-50 md:hover:translate-x-0.5 md:hover:shadow-[inset_3px_0_0_theme(colors.green.500)] active:bg-gray-200"
                  onClick={() => handleRowClick(ingredient)}
                >
                  <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs font-medium text-slate-700">{ingredient.name}</td>
                  <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">{ingredient.kcalPer100g}</td>
                  <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">{ingredient.proteins ?? '-'}</td>
                  <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">{ingredient.carbs ?? '-'}</td>
                  <td className="p-4 md:p-4 p-2.5 text-gray-800 text-sm md:text-sm text-xs">{ingredient.fats ?? '-'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {isEditModalOpen && selectedIngredient && (
        <EditIngredientModal
          ingredient={selectedIngredient}
          onUpdate={handleUpdate}
          onDelete={handleDelete}
          onCancel={() => {
            setIsEditModalOpen(false);
            setSelectedIngredient(null);
          }}
          isUpdating={updateMutation.isPending}
          isDeleting={deleteMutation.isPending}
        />
      )}

      {isAddModalOpen && (
        <AddIngredientModal
          onCreate={handleCreate}
          onCancel={() => setIsAddModalOpen(false)}
          isCreating={createMutation.isPending}
        />
      )}

      {showDeleteConfirmation && ingredientToDelete && (
        <DeleteConfirmationDialog
          message="foodList.confirmDelete"
          itemName={ingredientToDelete.name}
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
          isDeleting={deleteMutation.isPending}
        />
      )}
    </div>
  );
}