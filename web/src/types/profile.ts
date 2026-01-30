export interface ProfileData {
    id: number;
    first_name?: string;
    last_name?: string;
    email: string;
    age?: number;
    height?: number; // in centimeters
    weight?: number; // in kilograms    
    gender?: string;
    language: string; // en_US, uk_UA, ru_UA, bg_BG 
    activity_level?: string; // sedentary, lightly_active, moderate, very_active, extra_active
    created_at?: string;
    updated_at?: string;
    // Weight goal fields
    target_weight?: number; // in kilograms
    target_date?: string; // Format: YYYY-MM-DD, optional
    goal_set_at?: string;
    initial_weight_at_goal?: number; // in kilograms
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

