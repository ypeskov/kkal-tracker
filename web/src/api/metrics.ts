const API_BASE_URL = '/api';

export interface HealthMetrics {
  bmi?: number;
  bmi_category?: string;
  bmr?: number;
  tdee?: number;
  health_status?: string;
}

const getHeaders = () => {
  const token = sessionStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

export const metricsService = {
  getHealthMetrics: async (): Promise<HealthMetrics> => {
    const response = await fetch(`${API_BASE_URL}/metrics`, {
      method: 'GET',
      headers: getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch health metrics');
    }

    return response.json();
  },
};
