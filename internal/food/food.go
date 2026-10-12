// Package food holds the rules for food values shared by the diary and the ingredient list:
// sanity limits of calorie and nutrient values and the cleaning and comparison of food names.
package food

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Sanity limits: they catch unit mix-ups (calories of the whole portion instead of calories per 100 g,
// grams of a nutrient in the portion instead of in 100 g). Pure fat has 900 kcal per 100 g.
const (
	MaxKcalPer100g     = 900.0
	MaxNutrientPer100g = 100.0
	// MaxNameLength is counted in characters (runes), not bytes
	MaxNameLength = 200
)

func ValidKcalPer100g(kcalPer100g float64) bool {
	return kcalPer100g >= 0 && kcalPer100g <= MaxKcalPer100g
}

// ValidNutrientPer100g checks fats, carbs or proteins in grams per 100 g
func ValidNutrientPer100g(grams float64) bool {
	return grams >= 0 && grams <= MaxNutrientPer100g
}

// NameTooLong reports whether a (cleaned) food name exceeds MaxNameLength characters
func NameTooLong(name string) bool {
	return utf8.RuneCountInString(name) > MaxNameLength
}

// CleanName drops invisible characters (control and format ones, e.g. a zero-width space)
// and collapses whitespace, so that two names that look the same are the same
func CleanName(name string) string {
	visible := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, name)
	return strings.Join(strings.Fields(visible), " ")
}

// NormalizeName makes food names comparable: case, extra spaces, invisible characters and the Russian "ё" are ignored
func NormalizeName(name string) string {
	return strings.ReplaceAll(strings.ToLower(CleanName(name)), "ё", "е")
}
