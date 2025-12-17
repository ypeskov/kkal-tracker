import { aiService, AnalysisResult, AnalyzeRequest } from '@/api/ai';
import AIAnalysisPanel from '@/components/ai/AIAnalysisPanel';
import { useMutation, useQuery } from '@tanstack/react-query';
import { RefreshCw, Sparkles } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

type PeriodOption = 7 | 14 | 30 | 90 | 180 | 365;

export default function AIInsights() {
  const { t } = useTranslation();
  const [periodDays, setPeriodDays] = useState<PeriodOption>(7);
  const [analysisResult, setAnalysisResult] = useState<AnalysisResult | null>(null);
  const [analysisError, setAnalysisError] = useState<string | null>(null);

  const { data: status, isLoading: isLoadingStatus, isFetching: isFetchingStatus, refetch: refetchStatus } = useQuery({
    queryKey: ['aiStatus'],
    queryFn: () => aiService.getStatus(),
    staleTime: 0,
  });

  const analyzeMutation = useMutation({
    mutationFn: (request: AnalyzeRequest) => aiService.analyze(request),
    onSuccess: (result) => {
      setAnalysisResult(result);
      setAnalysisError(null);
    },
    onError: (error: Error) => {
      setAnalysisError(error.message);
      setAnalysisResult(null);
    },
  });

  const handleAnalyze = () => {
    if (!status?.available) return;

    analyzeMutation.mutate({
      period_days: periodDays,
    });
  };

  const periodOptions: { value: PeriodOption; label: string }[] = [
    { value: 7, label: t('ai.period.week') },
    { value: 14, label: t('ai.period.twoWeeks') },
    { value: 30, label: t('ai.period.month') },
    { value: 90, label: t('ai.period.threeMonths') },
    { value: 180, label: t('ai.period.sixMonths') },
    { value: 365, label: t('ai.period.year') },
  ];

  const isNotAvailable = !isLoadingStatus && !status?.available;

  return (
    <div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">
      <div className="mb-6">
        <h2 className="text-3xl font-semibold text-gray-800 flex items-center gap-2">
          <Sparkles className="text-blue-500" />
          {t('ai.title')}
        </h2>
        <p className="mt-2 text-gray-600">{t('ai.description')}</p>
      </div>

      {isNotAvailable ? (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 mb-6">
          <h3 className="text-lg font-medium text-yellow-800 mb-2">
            {t('ai.noProvidersTitle')}
          </h3>
          <p className="text-yellow-700">{t('ai.noProvidersDescription')}</p>
          <button
            onClick={() => refetchStatus()}
            disabled={isFetchingStatus}
            className="mt-4 px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 disabled:bg-yellow-400 transition-colors flex items-center gap-2"
          >
            <RefreshCw size={16} className={isFetchingStatus ? 'animate-spin' : ''} />
            {t('common.refresh')}
          </button>
        </div>
      ) : (
        <>
          {/* Controls */}
          <div className="bg-white rounded-lg shadow-md p-4 mb-6">
            <div className="flex flex-wrap gap-4 items-end">
              {/* Period Selector */}
              <div className="flex items-center gap-2">
                <label htmlFor="period" className="text-sm font-medium text-gray-700">
                  {t('ai.period.label')}:
                </label>
                <select
                  id="period"
                  value={periodDays}
                  onChange={(e) => setPeriodDays(Number(e.target.value) as PeriodOption)}
                  disabled={analyzeMutation.isPending}
                  className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 text-sm"
                >
                  {periodOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Analyze Button */}
              <button
                onClick={handleAnalyze}
                disabled={!status?.available || analyzeMutation.isPending}
                className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                <Sparkles size={18} />
                {t('ai.analyze')}
              </button>
            </div>
          </div>

          {/* Analysis Result */}
          <AIAnalysisPanel
            result={analysisResult}
            isLoading={analyzeMutation.isPending}
            error={analysisError}
          />
        </>
      )}
    </div>
  );
}
