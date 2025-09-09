import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { ingredientService, Ingredient, CreateIngredientData, UpdateIngredientData } from '../api/ingredients';
import EditIngredientModal from '../components/EditIngredientModal';
import AddIngredientModal from '../components/AddIngredientModal';
import DeleteConfirmationDialog from '../components/DeleteConfirmationDialog';
import './FoodList.css';

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

  // Filter ingredients based on search text
  const filteredIngredients = ingredients.filter(ingredient =>
    ingredient.name.toLowerCase().includes(filterText.toLowerCase())
  );

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
    return <div className="loading">{t('common.loading')}</div>;
  }

  if (error) {
    return <div className="error">{t('common.error')}</div>;
  }

  return (
    <div className="food-list-container">
      <div className="food-list-header">
        <h2>{t('foodList.title')}</h2>
        
        <div className="food-list-controls">
          <button 
            className="btn btn-primary"
            onClick={handleAddNew}
          >
            {t('foodList.addNew')}
          </button>
          
          <div className="filter-input">
            <input
              type="text"
              placeholder={t('foodList.filterPlaceholder')}
              value={filterText}
              onChange={(e) => setFilterText(e.target.value)}
              className="form-input"
            />
          </div>
        </div>
      </div>

      <div className="food-list-table-container">
        <table className="food-list-table">
          <thead>
            <tr>
              <th>{t('foodList.name')}</th>
              <th>{t('foodList.calories')}</th>
              <th>{t('foodList.proteins')}</th>
              <th>{t('foodList.carbs')}</th>
              <th>{t('foodList.fats')}</th>
            </tr>
          </thead>
          <tbody>
            {filteredIngredients.length === 0 ? (
              <tr>
                <td colSpan={5} className="no-data">
                  {t('foodList.noIngredients')}
                </td>
              </tr>
            ) : (
              filteredIngredients.map((ingredient) => (
                <tr
                  key={ingredient.id}
                  className="food-list-row"
                  onClick={() => handleRowClick(ingredient)}
                >
                  <td>{ingredient.name}</td>
                  <td>{ingredient.kcalPer100g}</td>
                  <td>{ingredient.proteins ?? '-'}</td>
                  <td>{ingredient.carbs ?? '-'}</td>
                  <td>{ingredient.fats ?? '-'}</td>
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