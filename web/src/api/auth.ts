interface LoginRequest {
  email: string
  password: string
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

  logout() {
    sessionStorage.removeItem('token')
  }

  getToken(): string | null {
    return sessionStorage.getItem('token')
  }
}

export const authService = new AuthService()