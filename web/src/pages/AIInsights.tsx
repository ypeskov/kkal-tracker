import { aiService, AnalysisResult, AnalyzeRequest } from '@/api/ai';
import AIAnalysisPanel from '@/components/ai/AIAnalysisPanel';
import AIProviderSelector from '@/components/ai/AIProviderSelector';
import { useMutation, useQuery } from '@tanstack/react-query';
import { RefreshCw, Send, Sparkles } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

type PeriodOption = 7 | 14 | 30 | 90 | 180 | 365;

export default function AIInsights() {
  const { t } = useTranslation();
  const [selectedProvider, setSelectedProvider] = useState<string>('');
  const [periodDays, setPeriodDays] = useState<PeriodOption>(7);
  const [query, setQuery] = useState<string>('');
  const [analysisResult, setAnalysisResult] = useState<AnalysisResult | null>(null);
  const [analysisError, setAnalysisError] = useState<string | null>(null);

  const { data: providers = [], isLoading: isLoadingProviders, isFetching: isFetchingProviders, refetch: refetchProviders } = useQuery({
    queryKey: ['aiProviders'],
    queryFn: () => aiService.getProviders(),
    staleTime: 0, // Always refetch on component mount
  });

  // Set default provider when providers are loaded
  if (providers.length > 0 && !selectedProvider) {
    setSelectedProvider(providers[0].id);
  }

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
    if (!selectedProvider) return;

    analyzeMutation.mutate({
      provider: selectedProvider,
      period_days: periodDays,
      query: query.trim() || undefined,
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

  const hasNoProviders = !isLoadingProviders && providers.length === 0;

  return (
    <div className="max-w-screen-xl mx-auto px-4 py-2 md:px-6 lg:px-8">
      <div className="mb-6">
        <h2 className="text-3xl font-semibold text-gray-800 flex items-center gap-2">
          <Sparkles className="text-blue-500" />
          {t('ai.title')}
        </h2>
        <p className="mt-2 text-gray-600">{t('ai.description')}</p>
      </div>

      {hasNoProviders ? (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 mb-6">
          <h3 className="text-lg font-medium text-yellow-800 mb-2">
            {t('ai.noProvidersTitle')}
          </h3>
          <p className="text-yellow-700">{t('ai.noProvidersDescription')}</p>
          <button
            onClick={() => refetchProviders()}
            disabled={isFetchingProviders}
            className="mt-4 px-4 py-2 bg-yellow-600 text-white rounded-md hover:bg-yellow-700 disabled:bg-yellow-400 transition-colors flex items-center gap-2"
          >
            <RefreshCw size={16} className={isFetchingProviders ? 'animate-spin' : ''} />
            {t('common.refresh')}
          </button>
        </div>
      ) : (
        <>
          {/* Controls */}
          <div className="bg-white rounded-lg shadow-md p-4 mb-6">
            <div className="flex flex-wrap gap-4 items-end">
              {/* Provider Selector */}
              <AIProviderSelector
                providers={providers}
                selected={selectedProvider}
                onChange={setSelectedProvider}
                disabled={analyzeMutation.isPending || isLoadingProviders}
              />

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
                disabled={!selectedProvider || analyzeMutation.isPending}
                className="px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
              >
                <Sparkles size={18} />
                {t('ai.analyze')}
              </button>
            </div>

            {/* Optional Question Input */}
            <div className="mt-4">
              <label htmlFor="query" className="block text-sm font-medium text-gray-700 mb-1">
                {t('ai.questionLabel')}
              </label>
              <div className="flex gap-2">
                <input
                  id="query"
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder={t('ai.questionPlaceholder')}
                  disabled={analyzeMutation.isPending}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && selectedProvider && !analyzeMutation.isPending) {
                      handleAnalyze();
                    }
                  }}
                />
                <button
                  onClick={handleAnalyze}
                  disabled={!selectedProvider || analyzeMutation.isPending}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:bg-gray-50 disabled:text-gray-400 disabled:cursor-not-allowed transition-colors"
                  title={t('ai.send')}
                >
                  <Send size={18} />
                </button>
              </div>
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




