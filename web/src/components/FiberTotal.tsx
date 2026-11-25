import { useTranslation } from 'react-i18next';
import { FiberTotals } from '@/utils/fiber';

interface FiberTotalProps {
  totals: FiberTotals;
  className?: string;
}

// Fibre total of a group of entries (a day or the filtered range):
// nothing when no entry has a known value, the grams otherwise (a known zero is shown as 0.0),
// with a visible hint when some entries have no fibre data, so the total is understated
export default function FiberTotal({ totals, className }: FiberTotalProps) {
  const { t } = useTranslation();

  if (totals.known === 0) return null;

  return (
    <span className={className}>
      {t('dashboard.fiberConsumed')}: {totals.grams.toFixed(1)} {t('common.g')}
      {totals.unknown > 0 && (
        <>
          {' '}
          <span className="font-normal text-amber-700">
            ({t('dashboard.fiberUnderstated', { count: totals.unknown })})
          </span>
        </>
      )}
    </span>
  );
}
