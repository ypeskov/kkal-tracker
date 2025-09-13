import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { authService } from '../api/auth'
import LanguageSwitcher from './LanguageSwitcher'

interface LoginProps {
  onLogin: () => void
}

export default function Login({ onLogin }: LoginProps) {
  const { t } = useTranslation()
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
      <LanguageSwitcher />
      <h2>{t('auth.login')} - {t('dashboard.title')}</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">{t('auth.email')}:</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="form-input"
            required
          />
        </div>
        
        <div className="form-group">
          <label htmlFor="password">{t('auth.password')}:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="form-input"
            required
          />
        </div>

        <button
          type="submit"
          className="btn btn--primary btn--full"
          disabled={loginMutation.isPending}
        >
          {loginMutation.isPending ? t('auth.loggingIn') : t('auth.loginButton')}
        </button>

        {loginMutation.error && (
          <div className="alert alert--error mt-md">
            {t('auth.loginFailed')}
          </div>
        )}
      </form>
    </div>
  )
}