// Dietary fibre helpers. Fibre (grams per 100 g) is optional: an absent value is unknown, which is not the same as 0

export const FIBER_MAX_PER_100G = 100;

// Value of the fibre field for a stored value: a known zero stays "0", unknown is an empty field
export const fiberToInput = (fiber?: number | null): string =>
  fiber === null || fiber === undefined ? '' : fiber.toString();

// Parses the fibre field: an empty field is unknown (undefined), a number within 0–100 is the value,
// anything else (out of range, an unfinished expression) is invalid (null)
export const parseFiberInput = (input: string): number | undefined | null => {
  const trimmed = input.trim();
  if (trimmed === '') return undefined;
  const value = Number(trimmed);
  if (!Number.isFinite(value) || value < 0 || value > FIBER_MAX_PER_100G) return null;
  return value;
};

export interface FiberTotals {
  // Grams of fibre in the entries with a known value
  grams: number;
  // Number of entries with a known value (0 included)
  known: number;
  // Number of entries without fibre data: when above zero, grams is understated
  unknown: number;
}

interface FiberEntry {
  fiber?: number | null;
  weight: number;
}

// Sums the fibre of a group of entries (a day or the filtered range)
export const sumFiber = (entries: FiberEntry[]): FiberTotals =>
  entries.reduce<FiberTotals>((totals, entry) => {
    if (entry.fiber === null || entry.fiber === undefined) {
      return { ...totals, unknown: totals.unknown + 1 };
    }
    return {
      ...totals,
      grams: totals.grams + (entry.fiber * entry.weight / 100),
      known: totals.known + 1,
    };
  }, { grams: 0, known: 0, unknown: 0 });
