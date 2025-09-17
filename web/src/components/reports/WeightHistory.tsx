import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { format } from 'date-fns';
import { weightService, WeightEntry } from '../../api/weight';
import { Pencil, Trash2, Plus, ArrowUp, ArrowDown } from 'lucide-react';

interface WeightHistoryProps {
  dateFrom?: string;
  dateTo?: string;
}

export default function WeightHistory({ dateFrom, dateTo }: WeightHistoryProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc'); // Default to descending (newest first)
  const [formData, setFormData] = useState({
    weight: '',
    recorded_at: format(new Date(), 'yyyy-MM-dd'),
  });

  // Fetch weight history
  const { data: weightHistory, isLoading } = useQuery({
    queryKey: ['weightHistory', dateFrom, dateTo],
    queryFn: () => weightService.getWeightHistory(dateFrom, dateTo),
  });

  // Sort weight history on frontend
  const sortedWeightHistory = useMemo(() => {
    if (!weightHistory) return [];

    return [...weightHistory].sort((a, b) => {
      const dateA = new Date(a.recorded_at);
      const dateB = new Date(b.recorded_at);

      if (sortOrder === 'desc') {
        return dateB.getTime() - dateA.getTime(); // Newest first
      } else {
        return dateA.getTime() - dateB.getTime(); // Oldest first
      }
    });
  }, [weightHistory, sortOrder]);

  // Create mutation
  const createMutation = useMutation({
    mutationFn: () => weightService.createWeightEntry({
      weight: parseFloat(formData.weight),
      recorded_at: formData.recorded_at,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['weightHistory'] });
      queryClient.invalidateQueries({ queryKey: ['reportData'] });
      setShowAddForm(false);
      setFormData({
        weight: '',
        recorded_at: format(new Date(), 'yyyy-MM-dd'),
      });
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, weight, recorded_at }: { id: number; weight: number; recorded_at: string }) =>
      weightService.updateWeightEntry(id, { weight, recorded_at }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['weightHistory'] });
      queryClient.invalidateQueries({ queryKey: ['reportData'] });
      setEditingId(null);
      setFormData({
        weight: '',
        recorded_at: format(new Date(), 'yyyy-MM-dd'),
      });
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => weightService.deleteWeightEntry(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['weightHistory'] });
      queryClient.invalidateQueries({ queryKey: ['reportData'] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingId) {
      updateMutation.mutate({
        id: editingId,
        weight: parseFloat(formData.weight),
        recorded_at: formData.recorded_at,
      });
    } else {
      createMutation.mutate();
    }
  };

  const handleEdit = (entry: WeightEntry) => {
    setEditingId(entry.id);
    setFormData({
      weight: entry.weight.toString(),
      recorded_at: format(new Date(entry.recorded_at), 'yyyy-MM-dd'),
    });
    setShowAddForm(true);
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setShowAddForm(false);
    setFormData({
      weight: '',
      recorded_at: format(new Date(), 'yyyy-MM-dd'),
    });
  };

  const toggleSortOrder = () => {
    setSortOrder(current => current === 'desc' ? 'asc' : 'desc');
  };

  if (isLoading) {
    return <div className="text-center py-8">{t('common.loading')}</div>;
  }

  return (
    <div className="space-y-4">
      {/* Add/Edit Form */}
      {showAddForm && (
        <div className="card p-4">
          <h3 className="font-semibold mb-3">
            {editingId ? t('report.edit_weight_entry') : t('report.add_weight_entry')}
          </h3>
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium mb-1">
                  {t('report.weight')} (kg)
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="1"
                  max="500"
                  value={formData.weight}
                  onChange={(e) => setFormData({ ...formData, weight: e.target.value })}
                  className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 py-3 rounded-lg shadow-sm"
                  placeholder="Enter weight"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  {t('report.date')}
                </label>
                <input
                  type="date"
                  value={formData.recorded_at}
                  onChange={(e) => setFormData({ ...formData, recorded_at: e.target.value })}
                  className="input border-2 bg-gray-50 focus:bg-white font-semibold text-lg px-4 py-3 rounded-lg shadow-sm"
                  required
                />
              </div>
            </div>
            <div className="flex gap-2">
              <button
                type="submit"
                className="btn-primary"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {editingId ? t('common.update') : t('common.save')}
              </button>
              <button
                type="button"
                onClick={handleCancelEdit}
                className="btn-secondary"
              >
                {t('common.cancel')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Add Button */}
      {!showAddForm && (
        <button
          onClick={() => setShowAddForm(true)}
          className="btn-primary flex items-center gap-2"
        >
          <Plus size={16} />
          {t('report.add_weight_entry')}
        </button>
      )}

      {/* Weight History Table */}
      <div className="card overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b">
              <th className="text-left p-3 font-medium">
                <button
                  onClick={toggleSortOrder}
                  className="flex items-center gap-1 hover:text-blue-600 transition-colors"
                >
                  {t('report.date')}
                  {sortOrder === 'desc' ? (
                    <ArrowDown size={16} className="text-blue-600" />
                  ) : (
                    <ArrowUp size={16} className="text-blue-600" />
                  )}
                </button>
              </th>
              <th className="text-left p-3 font-medium">{t('report.weight')}</th>
              <th className="text-right p-3 font-medium">{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {!sortedWeightHistory || sortedWeightHistory.length === 0 ? (
              <tr>
                <td colSpan={3} className="text-center p-8 text-gray-500">
                  {t('report.no_weight_entries')}
                </td>
              </tr>
            ) : (
              sortedWeightHistory.map((entry) => (
                <tr key={entry.id} className="border-b hover:bg-gray-50">
                  <td className="p-3">
                    {format(new Date(entry.recorded_at), 'PPP')}
                  </td>
                  <td className="p-3">
                    {entry.weight.toFixed(2)} kg
                  </td>
                  <td className="p-3 text-right">
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={() => handleEdit(entry)}
                        className="p-1.5 text-blue-600 hover:bg-blue-50 rounded"
                        title={t('common.edit')}
                      >
                        <Pencil size={16} />
                      </button>
                      <button
                        onClick={() => {
                          if (confirm(t('report.confirm_delete_weight'))) {
                            deleteMutation.mutate(entry.id);
                          }
                        }}
                        className="p-1.5 text-red-600 hover:bg-red-50 rounded"
                        title={t('common.delete')}
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}