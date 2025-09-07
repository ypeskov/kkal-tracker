import { useTranslation } from 'react-i18next'

const languages = [
  { code: 'en-US', name: 'language.en_US' },
  { code: 'uk-UA', name: 'language.uk_UA' },
  { code: 'ru-UA', name: 'language.ru_UA' }
]

export default function LanguageSwitcher() {
  const { t, i18n } = useTranslation()

  const changeLanguage = (languageCode: string) => {
    i18n.changeLanguage(languageCode)
  }

  return (
    <div className="language-switcher" style={{ margin: '1rem 0' }}>
      <label htmlFor="language-select" style={{ marginRight: '0.5rem' }}>
        {t('language.select')}:
      </label>
      <select
        id="language-select"
        value={i18n.language}
        onChange={(e) => changeLanguage(e.target.value)}
        style={{
          padding: '0.25rem',
          borderRadius: '4px',
          border: '1px solid #ccc'
        }}
      >
        {languages.map((language) => (
          <option key={language.code} value={language.code}>
            {t(language.name)}
          </option>
        ))}
      </select>
    </div>
  )
}