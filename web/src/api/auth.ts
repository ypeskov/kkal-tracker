interface LoginRequest {
  email: string
  password: string
}

interface User {
  id: number
  email: string
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
    localStorage.setItem('token', data.token)
    return data
  }

  async getCurrentUser(): Promise<User> {
    const token = localStorage.getItem('token')
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
    localStorage.removeItem('token')
  }
}

export const authService = new AuthService()