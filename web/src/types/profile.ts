export interface ProfileData {
    id: number;
    first_name?: string;
    last_name?: string;
    email: string;
    age?: number;
    height?: number;
    weight?: number;
    gender?: string;
    language: string;
    activity_level?: string;
    created_at?: string;
    updated_at?: string;
    // Weight goal fields
    target_weight?: number;
    target_date?: string;
    goal_set_at?: string;
    initial_weight_at_goal?: number;
}

export interface ProfileUpdateRequest {
    first_name?: string;
    last_name?: string;
    email: string;
    age?: number;
    height?: number;
    gender?: string;
    language: string;
    activity_level?: string;
}

export interface WeightGoalRequest {
    target_weight: number;
    target_date?: string; // Format: YYYY-MM-DD, optional
}

export interface WeightGoalProgress {
    target_weight: number;
    target_date?: string;
    goal_set_at: string;
    initial_weight_at_goal: number;
    current_weight: number;
    progress_percent: number;
    weight_to_go: number;
    weight_lost: number;
    days_remaining?: number;
    daily_deficit_needed?: number;
    estimated_completion?: string;
    is_gaining: boolean;
}

