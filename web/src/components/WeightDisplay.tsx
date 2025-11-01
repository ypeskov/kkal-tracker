import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Info } from 'lucide-react';

interface WeightDisplayProps {
  weightHistory: Array<{ weight: number; recorded_at: string }> | undefined;
}

export default function WeightDisplay({ weightHistory }: WeightDisplayProps) {
  const { t } = useTranslation();
  const [showTooltip, setShowTooltip] = useState(false);

  const latestWeight = weightHistory && weightHistory.length > 0
    ? weightHistory.sort((a, b) => new Date(b.recorded_at).getTime() - new Date(a.recorded_at).getTime())[0].weight
    : null;

  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">
        <span className="flex items-center gap-2">
          {t('profile.weight')}
          <div className="relative inline-block">
            <button
              type="button"
              className="text-gray-400 hover:text-gray-600 transition-colors"
              onMouseEnter={() => setShowTooltip(true)}
              onMouseLeave={() => setShowTooltip(false)}
              onClick={() => setShowTooltip(!showTooltip)}
            >
              <Info size={16} />
            </button>
            {showTooltip && (
              <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-2 bg-gray-800 text-white text-sm rounded-lg whitespace-nowrap z-10">
                <div className="absolute bottom-0 left-1/2 transform -translate-x-1/2 translate-y-1/2 rotate-45 w-2 h-2 bg-gray-800"></div>
                {t('profile.weightTooltip')}
              </div>
            )}
          </div>
        </span>
      </label>
      <div className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-gray-700">
        {latestWeight !== null ? (
          <span>{latestWeight.toFixed(1)} kg</span>
        ) : (
          <span className="text-gray-400">{t('profile.noWeightData')}</span>
        )}
      </div>
    </div>
  );
}
