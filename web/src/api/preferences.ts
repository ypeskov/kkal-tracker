import { ReportPreferences } from '@/types/preferences';

export type { ReportPreferences };

class PreferencesService {
  private getHeaders = () => {
    const token = sessionStorage.getItem('token');
    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  }

  // Returns the raw response body: the caller validates it field by field (useReportPreferences)
  getReportPreferences = async (): Promise<unknown> => {
    const response = await fetch('/api/preferences/reports', {
      method: 'GET',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to load report preferences');
    }

    return response.json();
  }

  // Replaces the whole stored value. keepalive lets a save started right before a reload or a tab close complete
  saveReportPreferences = async (preferences: ReportPreferences): Promise<void> => {
    const response = await fetch('/api/preferences/reports', {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(preferences),
      keepalive: true,
    });

    if (!response.ok) {
      throw new Error(`Failed to save report preferences: HTTP ${response.status}`);
    }
  }
}

export const preferencesService = new PreferencesService();
