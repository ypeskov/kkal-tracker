import { authService } from './auth';

const API_BASE_URL = '/api/weight';

export interface WeightEntry {
  id: number;
  user_id: number;
  weight: number;
  recorded_at: string;
  created_at: string;
}

export interface CreateWeightRequest {
  weight: number;
  recorded_at?: string;
}

export interface UpdateWeightRequest {
  weight: number;
  recorded_at?: string;
}

class WeightService {
  private getHeaders() {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authService.getToken()}`,
    };
  }

  async getWeightHistory(from?: string, to?: string): Promise<WeightEntry[]> {
    const params = new URLSearchParams();
    if (from) params.append('from', from);
    if (to) params.append('to', to);

    const queryString = params.toString();
    const url = `${API_BASE_URL}${queryString ? `?${queryString}` : ''}`;

    const response = await fetch(url, {
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch weight history');
    }

    return response.json();
  }

  async createWeightEntry(data: CreateWeightRequest): Promise<WeightEntry> {
    const response = await fetch(API_BASE_URL, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error('Failed to create weight entry');
    }

    return response.json();
  }

  async updateWeightEntry(id: number, data: UpdateWeightRequest): Promise<WeightEntry> {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error('Failed to update weight entry');
    }

    return response.json();
  }

  async deleteWeightEntry(id: number): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/${id}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to delete weight entry');
    }
  }
}

export const weightService = new WeightService();