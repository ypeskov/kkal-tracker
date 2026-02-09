export type ExportDataType = 'weight' | 'food' | 'both';
export type DeliveryType = 'download' | 'email';

export interface ExportRequest {
  date_from: string;
  date_to: string;
  data_type: ExportDataType;
  delivery_type: DeliveryType;
}

export interface ExportEmailResponse {
  message: string;
}

class ExportService {
  private getHeaders() {
    const token = sessionStorage.getItem('token');
    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  }

  async exportData(request: ExportRequest): Promise<Blob | ExportEmailResponse> {
    const response = await fetch('/api/export', {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Export failed' }));
      throw new Error(error.message || 'Export failed');
    }

    // Check if response is file (download) or JSON (email)
    const contentType = response.headers.get('Content-Type');
    if (contentType?.includes('application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')) {
      return response.blob();
    }
    return response.json();
  }

  // Helper to trigger download from blob
  downloadBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  }
}

export const exportService = new ExportService();
