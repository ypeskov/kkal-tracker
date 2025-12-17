import { authService } from './auth';

const API_BASE_URL = '/api/ai';

export interface AIProvider {
  id: string;
  display_name: string;
  model: string;
}

export interface ProvidersResponse {
  providers: AIProvider[];
}

export interface AnalyzeRequest {
  provider: string;
  period_days: number;
  query?: string;
}

export interface AnalysisResult {
  analysis: string;
  provider: string;
  model: string;
  tokens_used?: number;
  duration_ms: number;
}

class AIService {
  private getHeaders() {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authService.getToken()}`,
    };
  }

  async getProviders(): Promise<AIProvider[]> {
    const response = await fetch(`${API_BASE_URL}/providers`, {
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch AI providers');
    }

    const data: ProvidersResponse = await response.json();
    return data.providers;
  }

  async analyze(request: AnalyzeRequest): Promise<AnalysisResult> {
    const response = await fetch(`${API_BASE_URL}/analyze`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || 'AI analysis failed');
    }

    return response.json();
  }
}

export const aiService = new AIService();




