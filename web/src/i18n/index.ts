import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

// Translation files
import enUS from './locales/en_US.json'
import ukUA from './locales/uk_UA.json'
import ruUA from './locales/ru_UA.json'

const resources = {
  'en-US': {
    translation: enUS
  },
  'uk-UA': {
    translation: ukUA
  },
  'ru-UA': {
    translation: ruUA
  }
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en-US',
    debug: false,
    
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'i18nextLng',
      caches: ['localStorage']
    },

    interpolation: {
      escapeValue: false // React already escapes values
    }
  })

export default i18n