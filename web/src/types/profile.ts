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

