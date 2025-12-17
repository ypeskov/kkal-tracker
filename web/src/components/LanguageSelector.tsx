import { languagesService } from '@/api/languages';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

// Helper function to convert backend language codes to i18n format
const convertLanguageCode = (backendCode: string): string => {
  return backendCode.replace('_', '-');
};

interface LanguageSelectorProps {
  value: string;
  onChange: (language: string) => void;
  className?: string;
  disabled?: boolean;
}

export default function LanguageSelector({ value, onChange, className = '', disabled = false }: LanguageSelectorProps) {
  const { t, i18n } = useTranslation();

  const { data: languages = [] } = useQuery({
    queryKey: ['languages'],
    queryFn: languagesService.getLanguages,
    staleTime: Infinity, // Languages rarely change
  });

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newLang = e.target.value;
    onChange(newLang);
    // Convert backend language code to i18n format and update UI language immediately
    const i18nLangCode = convertLanguageCode(newLang);
    if (i18nLangCode !== i18n.language) {
      i18n.changeLanguage(i18nLangCode);
    }
  };

  return (
    <select
      value={value}
      onChange={handleChange}
      disabled={disabled || languages.length === 0}
      className={`px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-500 ${className}`}
    >
      {languages.map(lang => (
        <option key={lang.code} value={lang.code}>
          {t(`language.${lang.code}`)}
        </option>
      ))}
    </select>
  );
}