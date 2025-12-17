import { AIProvider } from '@/api/ai';
import { useTranslation } from 'react-i18next';

interface AIProviderSelectorProps {
  providers: AIProvider[];
  selected: string;
  onChange: (providerId: string) => void;
  disabled?: boolean;
}

export default function AIProviderSelector({
  providers,
  selected,
  onChange,
  disabled = false,
}: AIProviderSelectorProps) {
  const { t } = useTranslation();

  if (providers.length === 0) {
    return (
      <div className="text-sm text-gray-500 italic">
        {t('ai.noProviders')}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <label htmlFor="ai-provider" className="text-sm font-medium text-gray-700">
        {t('ai.provider')}:
      </label>
      <select
        id="ai-provider"
        value={selected}
        onChange={(e) => onChange(e.target.value)}
        disabled={disabled}
        className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:cursor-not-allowed text-sm"
      >
        {providers.map((provider) => (
          <option key={provider.id} value={provider.id}>
            {provider.display_name} ({provider.model})
          </option>
        ))}
      </select>
    </div>
  );
}




