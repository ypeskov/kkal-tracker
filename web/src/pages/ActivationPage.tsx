import { useEffect, useState } from 'react'
import { useParams } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { authService } from '@/api/auth'

interface ActivationPageProps {
  token?: string // Optional prop for direct rendering outside router
}

export default function ActivationPage({ token: tokenProp }: ActivationPageProps) {
  // Try to get token from router params, fallback to prop
  let token: string;
  try {
    const params = useParams({ from: '/activate/$token' });
    token = params.token || tokenProp || '';
  } catch {
    // If router context is not available, use prop
    token = tokenProp || '';
  }

  const { t } = useTranslation()
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    if (!token) {
      setStatus('error');
      setErrorMessage(t('auth.activationFailed'));
      return;
    }

    const activateAccount = async () => {
      try {
        await authService.activate(token)
        setStatus('success')
      } catch (error) {
        setStatus('error')
        setErrorMessage(error instanceof Error ? error.message : t('auth.activationFailed'))
      }
    }

    activateAccount()
  }, [token, t])

  if (status === 'loading') {
    return (
      <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center px-4">
        <div className="bg-white rounded-xl shadow-lg p-8 w-96 text-center">
          <div className="mb-6">
            <div className="w-16 h-16 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              {t('auth.activating')}
            </h1>
          </div>
        </div>
      </div>
    )
  }

  if (status === 'success') {
    return (
      <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center px-4">
        <div className="bg-white rounded-xl shadow-lg p-8 w-96 text-center">
          <div className="mb-6">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h1 className="text-2xl font-bold text-gray-900 mb-4">
              {t('auth.activationSuccessTitle')}
            </h1>
            <p className="text-gray-600 mb-6">
              {t('auth.activationSuccess')}
            </p>
          </div>

          <a
            href="/"
            className="w-full inline-block bg-blue-500 hover:bg-blue-600 !text-white hover:!text-white hover:no-underline font-medium py-3 px-4 rounded-lg transition-colors"
          >
            {t('auth.loginButton')}++++
          </a>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center justify-center px-4">
      <div className="bg-white rounded-xl shadow-lg p-8 w-96 text-center">
        <div className="mb-6">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {t('auth.activationFailed')}
          </h1>
          <p className="text-gray-600">
            {errorMessage}
          </p>
        </div>

        <a
          href="/"
          className="w-full inline-block bg-gray-500 hover:bg-gray-600 !text-white hover:!text-white hover:no-underline font-medium py-3 px-4 rounded-lg transition-colors"
        >
          {t('auth.loginLink')}
        </a>
      </div>
    </div>
  )
}
