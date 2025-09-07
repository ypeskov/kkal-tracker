import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { authService } from '../api/auth'

interface LoginProps {
  onLogin: () => void
}

export default function Login({ onLogin }: LoginProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const loginMutation = useMutation({
    mutationFn: authService.login,
    onSuccess: () => {
      onLogin()
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    loginMutation.mutate({ email, password })
  }

  return (
    <div className="login-container">
      <h2>Login to Calorie Tracker</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email:</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>

        <button 
          type="submit" 
          className="btn"
          disabled={loginMutation.isPending}
        >
          {loginMutation.isPending ? 'Logging in...' : 'Login'}
        </button>

        {loginMutation.error && (
          <div className="error">
            Login failed. Please check your credentials.
          </div>
        )}
      </form>
    </div>
  )
}