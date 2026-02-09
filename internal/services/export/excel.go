package export

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	"ypeskov/kkal-tracker/internal/i18n"
	"ypeskov/kkal-tracker/internal/models"
)

// ExcelGenerator creates Excel files from export data
type ExcelGenerator struct {
	t *i18n.Translator
}

// NewExcelGenerator creates a new Excel generator
func NewExcelGenerator() *ExcelGenerator {
	return &ExcelGenerator{
		t: i18n.GetTranslator(),
	}
}

// tr is a shortcut for translator.Get
func (g *ExcelGenerator) tr(lang, key string) string {
	return g.t.Get(lang, key)
}

// Generate creates an Excel file with the requested data
func (g *ExcelGenerator) Generate(
	weightData []*models.WeightHistory,
	foodData []*models.CalorieEntry,
	dataType ExportDataType,
	language string,
) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Track if we've created any sheets
	sheetsCreated := 0

	// Create Weight History sheet if requested
	if dataType == ExportWeight || dataType == ExportBoth {
		sheetName := g.tr(language, "export.sheets.weight")
		if sheetsCreated == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}
		sheetsCreated++

		if err := g.writeWeightSheet(f, sheetName, weightData, language); err != nil {
			return nil, fmt.Errorf("failed to write weight sheet: %w", err)
		}
	}

	// Create Food Entries sheet if requested
	if dataType == ExportFood || dataType == ExportBoth {
		sheetName := g.tr(language, "export.sheets.food")
		if sheetsCreated == 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}
		sheetsCreated++

		if err := g.writeFoodSheet(f, sheetName, foodData, language); err != nil {
			return nil, fmt.Errorf("failed to write food sheet: %w", err)
		}
	}

	// Delete default Sheet1 if it still exists and we created other sheets
	if sheetsCreated > 0 {
		sheetIndex, _ := f.GetSheetIndex("Sheet1")
		if sheetIndex >= 0 {
			f.DeleteSheet("Sheet1")
		}
	}

	// Write to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

func (g *ExcelGenerator) writeWeightSheet(f *excelize.File, sheetName string, data []*models.WeightHistory, lang string) error {
	// Write headers
	f.SetCellValue(sheetName, "A1", g.tr(lang, "export.columns.date"))
	f.SetCellValue(sheetName, "B1", g.tr(lang, "export.columns.weight_kg"))

	// Style headers (bold)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, "A1", "B1", headerStyle)

	// Write data
	for i, entry := range data {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), entry.RecordedAt.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), entry.Weight)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 15)

	return nil
}

func (g *ExcelGenerator) writeFoodSheet(f *excelize.File, sheetName string, data []*models.CalorieEntry, lang string) error {
	// Write headers
	columns := []struct {
		col   string
		key   string
		width float64
	}{
		{"A", "export.columns.date", 12},
		{"B", "export.columns.time", 10},
		{"C", "export.columns.food", 30},
		{"D", "export.columns.weight_g", 12},
		{"E", "export.columns.calories", 12},
		{"F", "export.columns.kcal_per100", 12},
		{"G", "export.columns.fats", 12},
		{"H", "export.columns.carbs", 12},
		{"I", "export.columns.proteins", 12},
	}

	for _, col := range columns {
		f.SetCellValue(sheetName, col.col+"1", g.tr(lang, col.key))
		f.SetColWidth(sheetName, col.col, col.col, col.width)
	}

	// Style headers (bold)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheetName, "A1", "I1", headerStyle)

	// Write data
	for i, entry := range data {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), entry.MealDatetime.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), entry.MealDatetime.Format("15:04"))
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), entry.Food)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), entry.Weight)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), entry.Calories)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), entry.KcalPer100g)

		// Handle optional macros
		if entry.Fats != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), *entry.Fats)
		}
		if entry.Carbs != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), *entry.Carbs)
		}
		if entry.Proteins != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), *entry.Proteins)
		}
	}

	return nil
}
