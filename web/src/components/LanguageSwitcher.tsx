import { useTranslation } from 'react-i18next'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { profileAPI } from '@/api/profile'

// Helper function to convert i18n language codes to backend format
const convertToBackendLanguageCode = (i18nCode: string): string => {
  return i18nCode.replace('-', '_');
};

const languages = [
  { code: 'en-US', name: 'language.en_US' },
  { code: 'uk-UA', name: 'language.uk_UA' },
  { code: 'ru-UA', name: 'language.ru_UA' },
  { code: 'bg-BG', name: 'language.bg_BG' }
]

export default function LanguageSwitcher() {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()

  // Update language mutation
  const updateLanguageMutation = useMutation({
    mutationFn: async (language: string) => {
      // Get current profile data
      const currentProfile = queryClient.getQueryData(['profile']) as any
      if (!currentProfile) {
        throw new Error('Profile data not available')
      }

      // Update only the language field
      const updateData = {
        first_name: currentProfile.first_name || '',
        last_name: currentProfile.last_name || '',
        email: currentProfile.email,
        age: currentProfile.age || undefined,
        height: currentProfile.height || undefined,
        weight: currentProfile.weight || undefined,
        language: convertToBackendLanguageCode(language),
      }

      return profileAPI.updateProfile(updateData)
    },
    onSuccess: (data) => {
      // Update profile query data
      queryClient.setQueryData(['profile'], data)

      // Update user query data as well
      queryClient.setQueryData(['user'], (oldUserData: any) => {
        if (oldUserData) {
          return { ...oldUserData, language: data.language }
        }
        return oldUserData
      })
    },
    onError: (error) => {
      console.error('Failed to save language preference:', error)
      // Revert UI language change on error
      const currentProfile = queryClient.getQueryData(['profile']) as any
      if (currentProfile?.language) {
        const revertLanguage = currentProfile.language.replace('_', '-')
        i18n.changeLanguage(revertLanguage)
      }
    }
  })

  const changeLanguage = (languageCode: string) => {
    // Change UI language immediately for better UX
    i18n.changeLanguage(languageCode)

    // Save to database
    updateLanguageMutation.mutate(languageCode)
  }

  return (
    <select
      id="language-select"
      value={i18n.language}
      onChange={(e) => changeLanguage(e.target.value)}
      disabled={updateLanguageMutation.isPending}
      className={`px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white text-sm ${updateLanguageMutation.isPending ? 'opacity-50 cursor-not-allowed' : ''}`}
      title={updateLanguageMutation.isPending ? 'Saving language preference...' : 'Select language'}
    >
      {languages.map((language) => (
        <option key={language.code} value={language.code}>
          {t(language.name)}
        </option>
      ))}
    </select>
  )
}