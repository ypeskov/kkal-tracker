package export

// Servicer defines the export service contract used by handlers.
type Servicer interface {
	Export(req *ExportRequest) (*ExportResult, error)
}
