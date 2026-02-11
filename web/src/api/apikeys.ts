export interface APIKey {
  id: number;
  name: string;
  key_prefix: string;
  expires_at: string | null;
  is_revoked: boolean;
  created_at: string;
}

export interface CreateAPIKeyRequest {
  name: string;
  expiry_days?: number;
}

export interface CreateAPIKeyResponse {
  id: number;
  name: string;
  key: string;
  key_prefix: string;
  expires_at: string | null;
  created_at: string;
}

class APIKeysService {
  private getHeaders() {
    const token = sessionStorage.getItem('token');
    return {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
    };
  }

  listKeys = async (): Promise<APIKey[]> => {
    const response = await fetch('/api/api-keys', {
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to list API keys' }));
      throw new Error(error.message || 'Failed to list API keys');
    }

    return response.json();
  };

  createKey = async (request: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    const response = await fetch('/api/api-keys', {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to create API key' }));
      throw new Error(error.message || 'Failed to create API key');
    }

    return response.json();
  };

  revokeKey = async (id: number): Promise<void> => {
    const response = await fetch(`/api/api-keys/${id}/revoke`, {
      method: 'POST',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to revoke API key' }));
      throw new Error(error.message || 'Failed to revoke API key');
    }
  };

  deleteKey = async (id: number): Promise<void> => {
    const response = await fetch(`/api/api-keys/${id}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Failed to delete API key' }));
      throw new Error(error.message || 'Failed to delete API key');
    }
  };
}

export const apiKeysService = new APIKeysService();
