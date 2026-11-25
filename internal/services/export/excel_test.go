package export

import (
	"bytes"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"ypeskov/kkal-tracker/internal/models"
)

func float(v float64) *float64 { return &v }

// The food sheet has the fibre per 100 g in column J: a value, a known zero, and an empty cell for an unknown one
func TestFoodSheetHasFiberColumn(t *testing.T) {
	when := time.Date(2026, 3, 10, 12, 30, 0, 0, time.UTC)
	entries := []*models.CalorieEntry{
		{Food: "Овсянка", Calories: 185, Weight: 50, KcalPer100g: 370, Fats: float(6), Carbs: float(60), Proteins: float(12),
			Fiber: float(10.6), MealDatetime: when},
		{Food: "Стейк", Calories: 250, Weight: 100, KcalPer100g: 250, Fiber: float(0), MealDatetime: when},
		{Food: "Суп", Calories: 60, Weight: 100, KcalPer100g: 60, Proteins: float(3), MealDatetime: when},
	}

	headers := map[string][]string{
		"en_US": {"Date", "Time", "Food", "Weight (g)", "Calories", "Kcal/100g", "Fats (g)", "Carbs (g)", "Proteins (g)", "Fiber (g)"},
		"uk_UA": {"", "", "", "", "", "", "", "", "", "Клітковина (г)"},
		"ru_UA": {"", "", "", "", "", "", "", "", "", "Клетчатка (г)"},
		"bg_BG": {"", "", "", "", "", "", "", "", "", "Фибри (г)"},
	}
	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	for language, want := range headers {
		t.Run(language, func(t *testing.T) {
			content, err := NewExcelGenerator().Generate(nil, entries, ExportFood, language)
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			file, err := excelize.OpenReader(bytes.NewReader(content))
			if err != nil {
				t.Fatalf("the export is not a readable XLSX file: %v", err)
			}
			defer file.Close()
			sheet := file.GetSheetList()[0]

			cell := func(ref string) string {
				t.Helper()
				value, err := file.GetCellValue(sheet, ref)
				if err != nil {
					t.Fatalf("%s: %v", ref, err)
				}
				return value
			}

			for i, column := range columns {
				header := cell(column + "1")
				if header == "" || header == "export.columns.fiber" || (want[i] != "" && header != want[i]) {
					t.Errorf("%s1 = %q, want %q (a translated header)", column, header, want[i])
				}
				// Every header is bold, the new one too
				styleID, err := file.GetCellStyle(sheet, column+"1")
				if err != nil {
					t.Fatalf("%s1 style: %v", column, err)
				}
				style, err := file.GetStyle(styleID)
				if err != nil || style.Font == nil || !style.Font.Bold {
					t.Errorf("%s1 is not bold (style %+v, error %v)", column, style, err)
				}
			}
			if beyond := cell("K1"); beyond != "" {
				t.Errorf("K1 = %q, want no eleventh column", beyond)
			}

			// Columns A-I are as before, J holds the fibre
			rows := [][]string{
				{"2026-03-10", "12:30", "Овсянка", "50", "185", "370", "6", "60", "12", "10.6"},
				{"2026-03-10", "12:30", "Стейк", "100", "250", "250", "", "", "", "0"},
				{"2026-03-10", "12:30", "Суп", "100", "60", "60", "", "", "3", ""},
			}
			for i, row := range rows {
				for j, column := range columns {
					ref := column + string(rune('2'+i))
					if got := cell(ref); got != row[j] {
						t.Errorf("%s = %q, want %q", ref, got, row[j])
					}
				}
			}
		})
	}
}
