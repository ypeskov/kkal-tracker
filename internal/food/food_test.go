package food

import (
	"strings"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		a, b string
		same bool
	}{
		{"Помидор", "помидор", true},
		{"  Сыр   Гауда ", "сыр гауда", true},
		{"Сёмга", "семга", true},
		{"Ёжик", "ежик", true},
		{"Jerky Pork", "JERKY PORK", true},
		{"Milk\u200b", "milk", true},
		{"Сыр\tГауда\n", "сыр гауда", true},
		{"Сыр Ементаль", "Сыр Ементаль плавленный", false},
		{"Помидор", "Помидор1", false},
	}
	for _, tt := range tests {
		if got := NormalizeName(tt.a) == NormalizeName(tt.b); got != tt.same {
			t.Errorf("NormalizeName(%q) == NormalizeName(%q): got %v, want %v", tt.a, tt.b, got, tt.same)
		}
	}
}

func TestCleanName(t *testing.T) {
	tests := map[string]string{
		"  Хлеб\u200b   ржаной ": "Хлеб ржаной",
		"Сыр\tГауда\n":           "Сыр Гауда",
		"\u200b \u00a0":          "",
		"Ёжик":                   "Ёжик", // case and "ё" are kept, only normalization ignores them
	}
	for name, want := range tests {
		if got := CleanName(name); got != want {
			t.Errorf("CleanName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestLimits(t *testing.T) {
	for _, tt := range []struct {
		value float64
		kcal  bool
		macro bool
	}{
		{-0.1, false, false},
		{0, true, true},
		{100, true, true},
		{100.1, true, false},
		{900, true, false},
		{900.1, false, false},
	} {
		if got := ValidKcalPer100g(tt.value); got != tt.kcal {
			t.Errorf("ValidKcalPer100g(%g) = %v, want %v", tt.value, got, tt.kcal)
		}
		if got := ValidNutrientPer100g(tt.value); got != tt.macro {
			t.Errorf("ValidNutrientPer100g(%g) = %v, want %v", tt.value, got, tt.macro)
		}
	}

	// The length is counted in characters: 200 Cyrillic letters are 400 bytes but still allowed
	if NameTooLong(strings.Repeat("я", MaxNameLength)) {
		t.Errorf("a name of %d characters must be allowed", MaxNameLength)
	}
	if !NameTooLong(strings.Repeat("я", MaxNameLength+1)) {
		t.Errorf("a name of %d characters must be rejected", MaxNameLength+1)
	}
}
