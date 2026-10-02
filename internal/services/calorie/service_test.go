package calorie

import "testing"

func TestNormalizeFoodName(t *testing.T) {
	tests := []struct {
		a, b string
		same bool
	}{
		{"Помидор", "помидор", true},
		{"  Сыр   Гауда ", "сыр гауда", true},
		{"Сёмга", "семга", true},
		{"Jerky Pork", "JERKY PORK", true},
		{"Milk\u200b", "milk", true},
		{"Сыр\tГауда\n", "сыр гауда", true},
		{"Сыр Ементаль", "Сыр Ементаль плавленный", false},
		{"Помидор", "Помидор1", false},
	}
	for _, tt := range tests {
		if got := normalizeFoodName(tt.a) == normalizeFoodName(tt.b); got != tt.same {
			t.Errorf("normalizeFoodName(%q) == normalizeFoodName(%q): got %v, want %v", tt.a, tt.b, got, tt.same)
		}
	}
}
