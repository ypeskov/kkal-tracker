import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, useMemo, useEffect } from 'react';
import { calorieService } from '../api/calories';
import { ingredientService } from '../api/ingredients';
import DashboardHeader from '../components/DashboardHeader';
import AddFoodEntryForm from '../components/AddFoodEntryForm';
import FilterSection from '../components/FilterSection';
import CalorieEntriesList from '../components/CalorieEntriesList';
import EditEntryModal from '../components/EditEntryModal';
import DeleteConfirmationDialog from '../components/DeleteConfirmationDialog';
import '../components/Dashboard.css';
import { useRouteContext } from '@tanstack/react-router';
import { RouterContext } from '../router';

type FilterType = 'today' | 'yesterday' | 'lastWeek' | 'lastMonth' | 'customRange';

export default function DashboardPage() {
  const { user, onLogout } = useRouteContext({ from: '__root__' }) as RouterContext;
  const [filterType, setFilterType] = useState<FilterType>('today');
  const [customDateFrom, setCustomDateFrom] = useState('');
  const [customDateTo, setCustomDateTo] = useState('');
  const [editingEntry, setEditingEntry] = useState<any>(null);
  const [showDeleteConfirmation, setShowDeleteConfirmation] = useState(false);
  const queryClient = useQueryClient();

  useEffect(() => {
    const loadIngredients = async () => {
      if (!ingredientService.hasCachedIngredients()) {
        try {
          await ingredientService.loadAndCacheIngredients();
        } catch (error) {
          console.error('Failed to load ingredients:', error);
        }
      }
    };
    loadIngredients();
  }, []);

  const getDateParams = useMemo(() => {
    const today = new Date().toISOString().split('T')[0];
    switch (filterType) {
      case 'today':
        return { dateFrom: today, dateTo: today };
      case 'yesterday':
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        const yesterdayStr = yesterday.toISOString().split('T')[0];
        return { dateFrom: yesterdayStr, dateTo: yesterdayStr };
      case 'lastWeek':
        const weekAgo = new Date();
        weekAgo.setDate(weekAgo.getDate() - 7);
        return { 
          dateFrom: weekAgo.toISOString().split('T')[0], 
          dateTo: today 
        };
      case 'lastMonth':
        const monthAgo = new Date();
        monthAgo.setDate(monthAgo.getDate() - 30);
        return { 
          dateFrom: monthAgo.toISOString().split('T')[0], 
          dateTo: today 
        };
      case 'customRange':
        if (customDateFrom && customDateTo) {
          return { dateFrom: customDateFrom, dateTo: customDateTo };
        }
        return { dateFrom: today, dateTo: today };
      default:
        return { dateFrom: today, dateTo: today };
    }
  }, [filterType, customDateFrom, customDateTo]);

  const { data: entries, isLoading } = useQuery({
    queryKey: ['calories', getDateParams],
    queryFn: () => calorieService.getEntries(getDateParams),
  });

  const nutritionTotals = useMemo(() => {
    if (!entries || entries.length === 0) {
      return { calories: 0, fats: 0, carbs: 0, proteins: 0 };
    }
    return entries.reduce((totals, entry) => ({
      calories: totals.calories + entry.calories,
      fats: totals.fats + ((entry.fats || 0) * entry.weight / 100),
      carbs: totals.carbs + ((entry.carbs || 0) * entry.weight / 100),
      proteins: totals.proteins + ((entry.proteins || 0) * entry.weight / 100)
    }), { calories: 0, fats: 0, carbs: 0, proteins: 0 });
  }, [entries]);

  const addEntryMutation = useMutation({
    mutationFn: calorieService.addEntry,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] });
    },
  });

  const updateEntryMutation = useMutation({
    mutationFn: ({ id, entry }: { id: number, entry: any }) => calorieService.updateEntry(id, entry),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] });
      setEditingEntry(null);
    },
  });

  const deleteEntryMutation = useMutation({
    mutationFn: (id: number) => calorieService.deleteEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calories'] });
      setEditingEntry(null);
      setShowDeleteConfirmation(false);
    },
  });

  const handleLogout = () => {
    ingredientService.clearCache();
    if (onLogout) {
      onLogout();
    }
  };

  const handleEdit = (entry: any) => {
    setEditingEntry(entry);
  };

  const handleCancelEdit = () => {
    setEditingEntry(null);
  };

  const handleShowDeleteConfirmation = () => {
    setShowDeleteConfirmation(true);
  };

  const handleConfirmDelete = () => {
    if (editingEntry) {
      deleteEntryMutation.mutate(editingEntry.id);
    }
  };

  const handleCancelDelete = () => {
    setShowDeleteConfirmation(false);
  };

  return (
    <div className="dashboard-container">
      <DashboardHeader user={user} onLogout={handleLogout} />

      <AddFoodEntryForm 
        onSubmit={addEntryMutation.mutate} 
        isSubmitting={addEntryMutation.isPending} 
      />

      <FilterSection 
        filterType={filterType}
        onFilterChange={setFilterType}
        customDateFrom={customDateFrom}
        onCustomDateFromChange={setCustomDateFrom}
        customDateTo={customDateTo}
        onCustomDateToChange={setCustomDateTo}
        nutritionTotals={nutritionTotals}
      />

      {isLoading ? (
        <p>Loading...</p>
      ) : (
        <CalorieEntriesList 
          entries={entries || []} 
          onEdit={handleEdit} 
        />
      )}

      {editingEntry && (
        <EditEntryModal
          entry={editingEntry}
          onUpdate={(id, entry) => updateEntryMutation.mutate({ id, entry })}
          onCancel={handleCancelEdit}
          onDelete={handleShowDeleteConfirmation}
          isUpdating={updateEntryMutation.isPending}
          isDeleting={deleteEntryMutation.isPending}
        />
      )}

      {showDeleteConfirmation && editingEntry && (
        <DeleteConfirmationDialog
          message="dashboard.deleteWarning"
          itemName={editingEntry.food}
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
          isDeleting={deleteEntryMutation.isPending}
        />
      )}
    </div>
  );
}