export interface ProfileData {
  id: number;
  first_name?: string;
  last_name?: string;
  email: string;
  age?: number;
  height?: number;
  weight?: number;
  language: string;
  created_at: string;
  updated_at: string;
}

export interface ProfileUpdateRequest {
  first_name?: string;
  last_name?: string;
  email: string;
  age?: number;
  height?: number;
  weight?: number;
  language: string;
}

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