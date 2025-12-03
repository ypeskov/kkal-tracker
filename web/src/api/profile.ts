import { ProfileData, ProfileUpdateRequest } from '@/types/profile';

export type { ProfileData, ProfileUpdateRequest };

class ProfileAPI {
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
}

export const profileAPI = new ProfileAPI();