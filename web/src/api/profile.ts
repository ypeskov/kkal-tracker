import { ProfileData, ProfileUpdateRequest, WeightGoalRequest, WeightGoalProgress } from '@/types/profile';

export type { ProfileData, ProfileUpdateRequest, WeightGoalRequest, WeightGoalProgress };

class ProfileService {
  private getHeaders = () => {
    const token = sessionStorage.getItem('token');
    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  }

  getProfile = async (): Promise<ProfileData> => {
    const response = await fetch('/api/profile', {
      method: 'GET',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch profile');
    }

    return response.json();
  }

  updateProfile = async (data: ProfileUpdateRequest): Promise<ProfileData> => {
    const response = await fetch('/api/profile', {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error('Failed to update profile');
    }

    return response.json();
  }

  setWeightGoal = async (data: WeightGoalRequest): Promise<WeightGoalProgress> => {
    const response = await fetch('/api/profile/goal', {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to set weight goal' }));
      throw new Error(error.message || 'Failed to set weight goal');
    }

    return response.json();
  }

  clearWeightGoal = async (): Promise<void> => {
    const response = await fetch('/api/profile/goal', {
      method: 'DELETE',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to clear weight goal');
    }
  }

  getWeightGoalProgress = async (): Promise<WeightGoalProgress | null> => {
    const response = await fetch('/api/profile/goal', {
      method: 'GET',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to get weight goal progress');
    }

    const data = await response.json();
    return data; // Can be null if no goal set
  }
}

export const profileService = new ProfileService();