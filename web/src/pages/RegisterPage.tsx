import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { authService } from '@/api/auth'
import LanguageSwitcher from '@/components/LanguageSwitcher'

export default function RegisterPage() {
  const { t, i18n } = useTranslation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  const registerMutation = useMutation({
    mutationFn: authService.register,
    onSuccess: () => {
      setSuccess(true)
      setError('')
    },
    onError: (err: Error) => {
      if (err.message.includes('already exists')) {
        setError(t('auth.userAlreadyExists'))
      } else {
        setError(t('auth.registrationFailed'))
      }
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Client-side validation
    if (password.length < 6) {
      setError(t('auth.passwordTooShort'))
      return
    }

    if (password !== confirmPassword) {
      setError(t('auth.passwordsDoNotMatch'))
      return
    }

    registerMutation.mutate({
      email,
      password,
      language_code: i18n.language.replace('-', '_'),
    })
  }

  if (success) {
    return (
      <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center px-4">
        <div className="absolute top-8">
          <LanguageSwitcher />
        </div>

        <div className="bg-white rounded-xl shadow-lg p-8 w-96 text-center">
          <div className="mb-6">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h1 className="text-2xl font-bold text-gray-900 mb-4">
              {t('auth.registrationSuccessTitle')}
            </h1>
            <p className="text-gray-600 mb-6">
              {t('auth.registrationSuccess')}
            </p>
          </div>

          <button
            onClick={() => window.location.href = '/'}
            className="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 px-4 rounded-lg transition-colors"
          >
            {t('auth.loginLink')}
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center px-4">
      <div className="absolute top-8">
        <LanguageSwitcher />
      </div>

      <div className="bg-white rounded-xl shadow-lg p-8 w-80">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            {t('auth.register')}
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
              minLength={6}
            />
          </div>

          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 mb-2">
              {t('auth.confirmPassword')}:
            </label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
              required
              minLength={6}
            />
          </div>

          <button
            type="submit"
            className="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-3 px-4 rounded-lg transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
            disabled={registerMutation.isPending}
          >
            {registerMutation.isPending ? t('auth.registering') : t('auth.registerButton')}
          </button>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-red-700 text-sm text-center">
              {error}
            </div>
          )}
        </form>

        <div className="mt-6 text-center text-sm">
          <span className="text-gray-600">{t('auth.alreadyHaveAccount')} </span>
          <a href="/" className="text-blue-500 hover:text-blue-600 font-medium">
            {t('auth.loginLink')}
          </a>
        </div>
      </div>
    </div>
  )
}
