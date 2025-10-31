interface LoginRequest {
  email: string
  password: string
}

interface RegisterRequest {
  email: string
  password: string
  language_code: string
}

interface RegisterResponse {
  message: string
  email: string
}

interface ActivationResponse {
  message: string
}

interface User {
  id: number
  email: string
  language?: string
}

class AuthService {
  async login(credentials: LoginRequest): Promise<{ token: string; user: User }> {
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(credentials),
    })

    if (!response.ok) {
      throw new Error('Login failed')
    }

    const data = await response.json()
    sessionStorage.setItem('token', data.token)
    return data
  }

  async getCurrentUser(): Promise<User> {
    const token = sessionStorage.getItem('token')
    if (!token) {
      throw new Error('No token found')
    }

    const response = await fetch('/api/auth/me', {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      throw new Error('Failed to get current user')
    }

    return response.json()
  }

  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await fetch('/api/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || 'Registration failed')
    }

    return response.json()
  }

  async activate(token: string): Promise<ActivationResponse> {
    const response = await fetch(`/api/auth/activate/${token}`, {
      method: 'GET',
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || 'Activation failed')
    }

    return response.json()
  }

  logout() {
    sessionStorage.removeItem('token')
  }

  getToken(): string | null {
    return sessionStorage.getItem('token')
  }
}

export const authService = new AuthService()