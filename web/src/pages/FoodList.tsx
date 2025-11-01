import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { ingredientService, Ingredient, CreateIngredientData, UpdateIngredientData } from '@/api/ingredients';
import EditIngredientModal from '@/components/EditIngredientModal';
import AddIngredientModal from '@/components/AddIngredientModal';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import PageHeader from '@/components/PageHeader';
import IngredientsTable from '@/components/IngredientsTable';
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
      <PageHeader
        title={t('foodList.title')}
        onAddNew={handleAddNew}
        addNewLabel={t('foodList.addNew')}
        filterValue={filterText}
        onFilterChange={setFilterText}
        filterPlaceholder={t('foodList.filterPlaceholder')}
      />

      <IngredientsTable
        ingredients={filteredIngredients}
        onRowClick={handleRowClick}
      />

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