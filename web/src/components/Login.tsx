import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { authService } from '@/api/auth'
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
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center px-4">
      {/* Language switcher at top */}
      <div className="absolute top-8">
        <LanguageSwitcher />
      </div>

      {/* Login card */}
      <div className="bg-white rounded-xl shadow-lg p-8 w-80">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            {t('auth.login')}
          </h1>
          <div className="w-8 h-px bg-gray-400 mx-auto mb-2"></div>
          <h2 className="text-lg text-gray-600">
            {t('dashboard.title')}
          </h2>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
              {t('auth.email')}:
            </label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              required
            />
          </div>

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-2">
              {t('auth.password')}:
            </label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              required
            />
          </div>

          <button
            type="submit"
            className="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 px-4 rounded-lg transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
            disabled={loginMutation.isPending}
          >
            {loginMutation.isPending ? t('auth.loggingIn') : t('auth.loginButton')}
          </button>

          {loginMutation.error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-red-700 text-sm text-center">
              {t('auth.loginFailed')}
            </div>
          )}
        </form>
      </div>
    </div>
  )
}