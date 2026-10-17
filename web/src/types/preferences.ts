export type ReportPeriod = 'daily' | 'weekly' | 'monthly' | 'yearly';
export type ReportStepInterval = '1' | '5' | '10' | '15'; // As the step <select> values

// Value of GET/PUT /api/preferences/reports
export interface ReportPreferences {
    date_from: string | null; // YYYY-MM-DD; null together with date_to = no saved range (rolling last 30 days)
    date_to: string | null; // YYYY-MM-DD or "today" (the browser's current day when the page is opened)
    show_weight: boolean;
    show_calories: boolean;
    show_goal_projection: boolean;
    period: ReportPeriod;
    step_interval: 1 | 5 | 10 | 15;
}

// The Reports page controls, with the range resolved to dates
export interface ReportSettings {
    dateFrom: string; // YYYY-MM-DD, '' while the input is cleared
    dateTo: string;
    showWeight: boolean;
    showCalories: boolean;
    showGoalProjection: boolean;
    period: ReportPeriod;
    stepInterval: ReportStepInterval;
}
