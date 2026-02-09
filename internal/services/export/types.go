package export

// ExportDataType defines what data to export
type ExportDataType string

const (
	ExportWeight ExportDataType = "weight"
	ExportFood   ExportDataType = "food"
	ExportBoth   ExportDataType = "both"
)

// DeliveryType defines how to deliver the export
type DeliveryType string

const (
	DeliveryDownload DeliveryType = "download"
	DeliveryEmail    DeliveryType = "email"
)

// ExportRequest contains all parameters for export
type ExportRequest struct {
	UserID       int
	DateFrom     string // YYYY-MM-DD
	DateTo       string // YYYY-MM-DD
	DataType     ExportDataType
	DeliveryType DeliveryType
	Language     string
	UserEmail    string // for email delivery
}

// ExportResult contains the generated file
type ExportResult struct {
	FileBytes []byte
	FileName  string
}
