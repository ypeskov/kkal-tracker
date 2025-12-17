import { authService } from './auth';

const API_BASE_URL = '/api/ai';

export interface AIStatus {
  available: boolean;
  model?: string;
}

export interface AnalyzeRequest {
  period_days: number;
}

export interface AnalysisResult {
  analysis: string;
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

  async getStatus(): Promise<AIStatus> {
    const response = await fetch(`${API_BASE_URL}/status`, {
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch AI status');
    }

    return response.json();
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
