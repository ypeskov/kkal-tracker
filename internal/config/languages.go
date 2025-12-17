package config

// Language represents a supported language
type Language struct {
	Code string // e.g., "en_US"
	Name string // e.g., "English"
}

// SupportedLanguages maps language codes to their full names
var SupportedLanguages = map[string]Language{
	"en_US": {Code: "en_US", Name: "English"},
	"uk_UA": {Code: "uk_UA", Name: "Ukrainian"},
	"ru_UA": {Code: "ru_UA", Name: "Russian"},
	"bg_BG": {Code: "bg_BG", Name: "Bulgarian"},
}

// DefaultLanguageCode is the fallback language
const DefaultLanguageCode = "en_US"

// GetLanguageName returns the full language name for a code, or English as default
func GetLanguageName(code string) string {
	if lang, ok := SupportedLanguages[code]; ok {
		return lang.Name
	}
	return SupportedLanguages[DefaultLanguageCode].Name
}

// GetSupportedLanguageCodes returns a slice of all supported language codes
func GetSupportedLanguageCodes() []string {
	codes := make([]string, 0, len(SupportedLanguages))
	for code := range SupportedLanguages {
		codes = append(codes, code)
	}
	return codes
}
