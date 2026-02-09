package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

const fallbackLanguage = "en_US"

// Translator provides access to localized strings
type Translator struct {
	translations map[string]map[string]any // language -> parsed JSON
	mu           sync.RWMutex
}

var (
	instance *Translator
	once     sync.Once
)

// GetTranslator returns the singleton translator instance
func GetTranslator() *Translator {
	once.Do(func() {
		instance = &Translator{
			translations: make(map[string]map[string]any),
		}
		instance.loadAll()
	})
	return instance
}

// loadAll reads all JSON files from the embedded locales directory
func (t *Translator) loadAll() {
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := localesFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			continue
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			continue
		}

		// Extract language from filename: en_US.json -> en_US
		lang := strings.TrimSuffix(entry.Name(), ".json")
		t.translations[lang] = parsed
	}
}

// Get retrieves a translated string by dot-separated key path.
// Falls back to en_US if the key is not found in the requested language.
// Returns the key itself if not found in any language.
func (t *Translator) Get(language, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try requested language
	if val := t.resolve(language, key); val != "" {
		return val
	}

	// Fallback to default language
	if language != fallbackLanguage {
		if val := t.resolve(fallbackLanguage, key); val != "" {
			return val
		}
	}

	return key
}

// resolve looks up a dot-separated key in a specific language
func (t *Translator) resolve(language, key string) string {
	langData, ok := t.translations[language]
	if !ok {
		return ""
	}

	parts := strings.Split(key, ".")
	var current any = langData

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = m[part]
		if !ok {
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	return fmt.Sprintf("%v", current)
}
