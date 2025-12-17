import { AnalysisResult } from '@/api/ai';
import { Clock, Coins, Sparkles } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface AIAnalysisPanelProps {
  result: AnalysisResult | null;
  isLoading: boolean;
  error: string | null;
}

export default function AIAnalysisPanel({
  result,
  isLoading,
  error,
}: AIAnalysisPanelProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-center py-12">
          <div className="flex flex-col items-center gap-4">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
            <p className="text-gray-600">{t('ai.analyzing')}</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-700">{error}</p>
        </div>
      </div>
    );
  }

  if (!result) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex flex-col items-center justify-center py-12 text-gray-500">
          <Sparkles size={48} className="mb-4 text-gray-300" />
          <p className="text-center">{t('ai.selectAndAnalyze')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center gap-2 mb-4">
        <Sparkles size={20} className="text-blue-500" />
        <h3 className="text-lg font-semibold text-gray-800">{t('ai.analysisResult')}</h3>
      </div>

      <div
        className="prose prose-sm max-w-none mb-4 text-gray-700 leading-relaxed [&>h3]:text-lg [&>h3]:font-semibold [&>h3]:text-gray-800 [&>h3]:mt-4 [&>h3]:mb-2 [&>p]:mb-3 [&>ul]:list-disc [&>ul]:pl-5 [&>ul]:mb-3 [&>ol]:list-decimal [&>ol]:pl-5 [&>ol]:mb-3"
        dangerouslySetInnerHTML={{ __html: result.analysis }}
      />

      <div className="flex flex-wrap gap-4 pt-4 border-t border-gray-100 text-sm text-gray-500">
        <div className="flex items-center gap-1">
          <Clock size={14} />
          <span>{(result.duration_ms / 1000).toFixed(1)}s</span>
        </div>
        {result.tokens_used && result.tokens_used > 0 && (
          <div className="flex items-center gap-1">
            <Coins size={14} />
            <span>{result.tokens_used} tokens</span>
          </div>
        )}
      </div>
    </div>
  );
}




