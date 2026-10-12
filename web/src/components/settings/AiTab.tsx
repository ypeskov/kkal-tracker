import { skillsService, SkillDownloadError } from '@/api/skills';
import NotificationPopup from '@/components/NotificationPopup';
import { downloadBlob } from '@/utils/download';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Download, TriangleAlert } from 'lucide-react';
import { useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';

const codeClassName = 'bg-gray-100 px-1.5 py-0.5 rounded text-gray-800 font-mono text-xs';

export default function AiTab() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const downloadMutation = useMutation({
    mutationFn: skillsService.downloadFoodLogSkill,
    onSuccess: ({ blob, filename }) => {
      downloadBlob(blob, filename);
      setNotification({ type: 'success', message: t('settings.ai.downloadSuccess') });
    },
    onError: (error: Error) => {
      // Fixed localized messages only: the server's response text is never shown
      const tooManyRequests = error instanceof SkillDownloadError && error.status === 429;
      setNotification({
        type: 'error',
        message: tooManyRequests ? t('settings.ai.tooManyRequests') : t('settings.ai.error'),
      });
    },
    // Also after a failure: the key may already exist if the response broke off mid-transfer
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] });
    },
  });

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <h3 className="text-xl font-semibold text-gray-800 mb-2">{t('settings.ai.title')}</h3>
      <p className="text-gray-600 mb-6">{t('settings.ai.description')}</p>

      <div className="space-y-6">
        {/* Download */}
        <div className="space-y-3">
          <button
            onClick={() => downloadMutation.mutate()}
            disabled={downloadMutation.isPending}
            className="inline-flex items-center gap-2 px-6 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            <Download size={18} />
            {downloadMutation.isPending ? t('settings.ai.preparing') : t('settings.ai.download')}
          </button>
          <p className="text-sm text-gray-600">{t('settings.ai.keyNote')}</p>
        </div>

        {/* Install steps */}
        <div>
          <h4 className="text-lg font-semibold text-gray-800 mb-3">{t('settings.ai.install.title')}</h4>
          <ol className="ml-5 list-decimal space-y-2 text-sm text-gray-700">
            <li>
              <Trans
                i18nKey="settings.ai.install.step1"
                values={{ domain: window.location.host }}
                components={{ code: <code className={codeClassName} /> }}
              />
            </li>
            <li>{t('settings.ai.install.step2')}</li>
            <li>{t('settings.ai.install.step3')}</li>
          </ol>
          <p className="mt-3 text-sm text-gray-700">
            <Trans
              i18nKey="settings.ai.install.claudeCode"
              values={{ path: '~/.claude/skills/' }}
              components={{ code: <code className={codeClassName} /> }}
            />
          </p>
        </div>

        {/* Secret key warning */}
        <div className="flex items-start gap-2 text-sm text-amber-600">
          <TriangleAlert size={18} className="flex-shrink-0 mt-0.5" />
          <p>{t('settings.ai.warning')}</p>
        </div>
      </div>

      {notification && (
        <NotificationPopup
          type={notification.type}
          message={notification.message}
          onClose={() => setNotification(null)}
        />
      )}
    </div>
  );
}
