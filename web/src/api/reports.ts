import { authService } from './auth';

const API_BASE_URL = '/api/reports';

export interface WeightDataPoint {
  date: string;
  weight: number;
}

export interface CalorieDataPoint {
  date: string;
  calories: number;
}

export interface ReportData {
  weight_history: WeightDataPoint[];
  calorie_history: CalorieDataPoint[];
}

class ReportsService {
  private getHeaders() {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authService.getToken()}`,
    };
  }

  async getReportData(from?: string, to?: string): Promise<ReportData> {
    const params = new URLSearchParams();
    if (from) params.append('from', from);
    if (to) params.append('to', to);

    const queryString = params.toString();
    const url = `${API_BASE_URL}/data${queryString ? `?${queryString}` : ''}`;

    const response = await fetch(url, {
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch report data');
    }

    return response.json();
  }
}

export const reportsService = new ReportsService();