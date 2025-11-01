import { useTranslation } from 'react-i18next';

interface PageHeaderProps {
  title: string;
  onAddNew?: () => void;
  addNewLabel?: string;
  filterValue?: string;
  onFilterChange?: (value: string) => void;
  filterPlaceholder?: string;
}

export default function PageHeader({
  title,
  onAddNew,
  addNewLabel,
  filterValue,
  onFilterChange,
  filterPlaceholder,
}: PageHeaderProps) {
  const { t } = useTranslation();

  return (
    <div className="mb-8">
      <h2 className="mb-5 text-gray-800 text-3xl font-semibold">{title}</h2>

      <div className="flex flex-col md:flex-row gap-5 items-stretch md:items-center mb-5">
        {onAddNew && addNewLabel && (
          <button
            className="btn-primary px-5 py-2.5 text-sm font-medium uppercase tracking-wide transition-all duration-300 hover:-translate-y-0.5 hover:shadow-lg disabled:opacity-60 disabled:cursor-not-allowed"
            onClick={onAddNew}
          >
            {addNewLabel}
          </button>
        )}

        {onFilterChange !== undefined && filterValue !== undefined && (
          <div className="relative flex-1">
            <input
              type="text"
              placeholder={filterPlaceholder}
              value={filterValue}
              onChange={(e) => onFilterChange(e.target.value)}
              className="w-full py-2.5 pl-4 pr-12 border border-gray-300 rounded text-sm transition-colors duration-300 focus:outline-none focus:border-green-500 placeholder-gray-500"
            />
            {filterValue && (
              <button
                type="button"
                className="absolute right-1 top-1/2 -translate-y-1/2 flex items-center justify-center p-2.5 text-gray-500 text-2xl leading-none transition-colors duration-200 hover:text-gray-700 active:text-gray-800"
                onClick={() => onFilterChange('')}
                aria-label={t('common.clear')}
              >
                ×
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
