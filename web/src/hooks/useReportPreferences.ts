import { authService } from '@/api/auth';
import { preferencesService } from '@/api/preferences';
import { RouterContext } from '@/router';
import { ReportPeriod, ReportPreferences, ReportSettings, ReportStepInterval } from '@/types/preferences';
import { noop, QueryClient, queryOptions, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouteContext } from '@tanstack/react-router';
import { format, subDays } from 'date-fns';
import { useEffect, useState } from 'react';

// One PUT after this long without further changes
const SAVE_DEBOUNCE_MS = 800;
// Range without a saved one: the last 30 days up to today, rolling
const DEFAULT_RANGE_DAYS = 30;
// Same bounds as the server
const MIN_DATE = '1900-01-01';
const MAX_DATE = '2100-12-31';

const PERIODS: readonly ReportPeriod[] = ['daily', 'weekly', 'monthly', 'yearly'];
const STEP_INTERVALS: readonly ReportPreferences['step_interval'][] = [1, 5, 10, 15];

export const DEFAULT_REPORT_PREFERENCES: ReportPreferences = {
  date_from: null,
  date_to: null,
  show_weight: true,
  show_calories: true,
  show_goal_projection: true,
  period: 'daily',
  step_interval: 1,
};

type SavedRange = Pick<ReportPreferences, 'date_from' | 'date_to'>;

const toDateString = (date: Date) => format(date, 'yyyy-MM-dd');

// A calendar date YYYY-MM-DD within the server's bounds
function isValidDate(value: unknown): value is string {
  if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  if (value < MIN_DATE || value > MAX_DATE) return false;
  const [year, month, day] = value.split('-').map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day;
}

const isPeriod = (value: unknown): value is ReportPeriod => PERIODS.includes(value as ReportPeriod);
const isStepInterval = (value: unknown): value is ReportPreferences['step_interval'] =>
  STEP_INTERVALS.includes(value as ReportPreferences['step_interval']);

// Validates a GET response with the server's rules: an invalid field gets its default, an invalid range resets both
// dates
export function parseReportPreferences(raw: unknown): ReportPreferences {
  const value = typeof raw === 'object' && raw !== null && !Array.isArray(raw) ? raw as Record<string, unknown> : {};
  const defaults = DEFAULT_REPORT_PREFERENCES;
  const rangeValid = isValidDate(value.date_from) && (value.date_to === 'today' || isValidDate(value.date_to));
  const boolOr = (field: unknown, fallback: boolean) => typeof field === 'boolean' ? field : fallback;

  return {
    date_from: rangeValid ? value.date_from as string : null,
    date_to: rangeValid ? value.date_to as string : null,
    show_weight: boolOr(value.show_weight, defaults.show_weight),
    show_calories: boolOr(value.show_calories, defaults.show_calories),
    show_goal_projection: boolOr(value.show_goal_projection, defaults.show_goal_projection),
    period: isPeriod(value.period) ? value.period : defaults.period,
    step_interval: isStepInterval(value.step_interval) ? value.step_interval : defaults.step_interval,
  };
}

// Saved preferences + today → the page controls. Without a saved range: today − 30 days … today; "today" → today
export function resolveReportSettings(preferences: ReportPreferences, today: Date): ReportSettings {
  const hasRange = preferences.date_from !== null && preferences.date_to !== null;
  const todayString = toDateString(today);

  return {
    dateFrom: hasRange ? preferences.date_from! : toDateString(subDays(today, DEFAULT_RANGE_DAYS)),
    dateTo: !hasRange || preferences.date_to === 'today' ? todayString : preferences.date_to!,
    showWeight: preferences.show_weight,
    showCalories: preferences.show_calories,
    showGoalProjection: preferences.show_goal_projection,
    period: preferences.period,
    stepInterval: String(preferences.step_interval) as ReportStepInterval,
  };
}

// The range as it is saved after the user changed a date: From fixed, To "today" when it is today. A To the user did
// not change stays "today" when it was saved so or not saved at all (the rolling default), even if the page was opened
// on an earlier day. null when an input is cleared or holds a date the server does not accept: such a range is not
// saved
export function toSavedRange(
  dateFrom: string,
  dateTo: string,
  previous: SavedRange,
  toChanged: boolean,
  today: Date,
): SavedRange | null {
  if (!isValidDate(dateFrom) || !isValidDate(dateTo)) return null;
  const keepsToday = !toChanged && (previous.date_to === 'today' || previous.date_to === null);
  return { date_from: dateFrom, date_to: keepsToday || dateTo === toDateString(today) ? 'today' : dateTo };
}

function toReportPreferences(settings: ReportSettings, range: SavedRange): ReportPreferences {
  return {
    ...range,
    show_weight: settings.showWeight,
    show_calories: settings.showCalories,
    show_goal_projection: settings.showGoalProjection,
    period: settings.period,
    step_interval: Number(settings.stepInterval) as ReportPreferences['step_interval'],
  };
}

// Debounced saves with at most one PUT in flight. Module-level, so a save flushed when the page unmounts still
// completes, and a later visit of the page queues behind it instead of racing it
class ReportPreferencesSaver {
  // Incremented on every change: lets a GET that was out during a change see that its response is older
  generation = 0;
  private pending: { preferences: ReportPreferences; token: string | null } | null = null;
  private timer: ReturnType<typeof setTimeout> | null = null;
  private inFlight = 0;
  private listening = false;

  schedule(preferences: ReportPreferences) {
    this.listen();
    this.generation++;
    this.pending = { preferences, token: authService.getToken() };
    if (this.timer !== null) clearTimeout(this.timer);
    this.timer = setTimeout(() => {
      this.timer = null;
      this.flush();
    }, SAVE_DEBOUNCE_MS);
  }

  // A change is waiting or being saved: the server does not have the latest state yet
  isBusy() {
    return this.pending !== null || this.inFlight > 0;
  }

  // Sends the pending change now. While a save is in flight it goes out after that one, unless force (the page is
  // being unloaded and would not live to send it)
  flush(force = false) {
    if (this.timer !== null) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    if (this.pending === null || (this.inFlight > 0 && !force)) return;

    const { preferences, token } = this.pending;
    this.pending = null;
    // Only for the session that made the change: after a logout or a login as someone else the change is dropped
    if (token === null || token !== authService.getToken()) return;

    this.inFlight++;
    preferencesService.saveReportPreferences(preferences)
      // Silent for the user: the page keeps its state, the next change sends the whole state again
      .catch((error: unknown) => console.warn('Report preferences were not saved', error))
      .finally(() => {
        this.inFlight--;
        // A change made meanwhile goes out now, unless its debounce is still running
        if (this.timer === null) this.flush();
      });
  }

  // Best effort on reload / tab close / app switch on mobile; the PUT uses keepalive, so it outlives the page
  private listen() {
    if (this.listening) return;
    this.listening = true;
    window.addEventListener('pagehide', () => this.flush(true));
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') this.flush();
    });
  }
}

const saver = new ReportPreferencesSaver();

// The user id keeps the cached settings of different accounts apart, besides the cache reset on logout and login
export const reportPreferencesQueryKey = (userId: number) => ['preferences', 'reports', userId] as const;

export const reportPreferencesQueryOptions = (userId: number) => queryOptions({
  queryKey: reportPreferencesQueryKey(userId),
  queryFn: async ({ client, queryKey }): Promise<ReportPreferences> => {
    const generation = saver.generation;
    const preferences = parseReportPreferences(await preferencesService.getReportPreferences());
    // A change made in this tab that the server may not have yet is newer than the response: keep it in the cache
    if (generation !== saver.generation || saver.isBusy()) {
      return client.getQueryData(queryKey) ?? preferences;
    }
    return preferences;
  },
  retry: 1,
});

// Loads the settings once after login (app shell), so the Reports page usually opens without waiting
export function prefetchReportPreferences(client: QueryClient, userId: number) {
  void client.query(reportPreferencesQueryOptions(userId)).catch(noop);
}

interface PageState {
  settings: ReportSettings;
  // The range as saved: null dates (rolling default) stay null until the user changes a date
  savedRange: SavedRange;
}

const initialPageState = (preferences: ReportPreferences): PageState => ({
  settings: resolveReportSettings(preferences, new Date()),
  savedRange: { date_from: preferences.date_from, date_to: preferences.date_to },
});

/**
 * The Reports page controls, restored from the user's saved preferences and saved back on every change.
 * settings is null until the saved preferences are loaded (or failed to load: then the defaults apply). They are
 * applied once per mount: a later refetch does not change the controls.
 */
export function useReportPreferences(): {
  settings: ReportSettings | null;
  update: (change: Partial<ReportSettings>) => void;
} {
  const { user } = useRouteContext({ from: '__root__' }) as RouterContext;
  const userId = user?.id;
  const queryClient = useQueryClient();

  const query = useQuery({
    ...reportPreferencesQueryOptions(userId ?? 0),
    enabled: userId !== undefined,
  });
  const loaded = query.data !== undefined || query.isError;

  const [state, setState] = useState<PageState | null>(
    () => query.data !== undefined ? initialPageState(query.data) : null,
  );
  // Settings that arrive after the first render: adopt them once (a failed GET → the defaults)
  if (state === null && loaded) {
    setState(initialPageState(query.data ?? DEFAULT_REPORT_PREFERENCES));
  }

  // Leaving the page sends a pending change at once
  useEffect(() => () => saver.flush(), []);

  const update = (change: Partial<ReportSettings>) => {
    if (state === null || userId === undefined) return;

    const settings = { ...state.settings, ...change };
    let savedRange = state.savedRange;
    if (change.dateFrom !== undefined || change.dateTo !== undefined) {
      const today = new Date();
      const range = toSavedRange(settings.dateFrom, settings.dateTo, savedRange, change.dateTo !== undefined, today);
      // A cleared or unacceptable date keeps the last saved range
      if (range !== null) {
        savedRange = range;
        // To "today" shows today's date: it moves forward when the page stayed open past midnight
        if (range.date_to === 'today') settings.dateTo = toDateString(today);
      }
    }
    setState({ settings, savedRange });

    const previous = toReportPreferences(state.settings, state.savedRange);
    const next = toReportPreferences(settings, savedRange);
    if (JSON.stringify(previous) === JSON.stringify(next)) return;

    // The cache holds the latest state at once, so coming back to the page shows it even before the save completes
    queryClient.setQueryData(reportPreferencesQueryKey(userId), next);
    saver.schedule(next);
  };

  return { settings: state?.settings ?? null, update };
}
