package export

import (
	"fmt"
	"log/slog"

	"ypeskov/kkal-tracker/internal/models"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	emailservice "ypeskov/kkal-tracker/internal/services/email"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"
)

// Service handles data export operations
type Service struct {
	calorieService *calorieservice.Service
	weightService  *weightservice.Service
	emailService   *emailservice.Service
	excelGenerator *ExcelGenerator
	logger         *slog.Logger
}

// New creates a new export service
func New(
	calorieService *calorieservice.Service,
	weightService *weightservice.Service,
	emailService *emailservice.Service,
	logger *slog.Logger,
) *Service {
	return &Service{
		calorieService: calorieService,
		weightService:  weightService,
		emailService:   emailService,
		excelGenerator: NewExcelGenerator(),
		logger:         logger.With("service", "export"),
	}
}

// Export generates an export file and either returns it or sends via email
func (s *Service) Export(req *ExportRequest) (*ExportResult, error) {
	s.logger.Debug("Export called",
		"user_id", req.UserID,
		"date_from", req.DateFrom,
		"date_to", req.DateTo,
		"data_type", req.DataType,
		"delivery_type", req.DeliveryType,
	)

	// Fetch weight data if needed
	var weightData []*models.WeightHistory
	if req.DataType == ExportWeight || req.DataType == ExportBoth {
		data, err := s.weightService.GetWeightHistoryByDateRange(req.UserID, req.DateFrom, req.DateTo)
		if err != nil {
			s.logger.Error("Failed to fetch weight data", "error", err)
			return nil, fmt.Errorf("failed to fetch weight data: %w", err)
		}
		weightData = data
		s.logger.Debug("Fetched weight data", "count", len(weightData))
	}

	// Fetch calorie data if needed
	var calorieData []*models.CalorieEntry
	if req.DataType == ExportFood || req.DataType == ExportBoth {
		data, err := s.calorieService.GetEntriesByDateRange(req.UserID, req.DateFrom, req.DateTo)
		if err != nil {
			s.logger.Error("Failed to fetch calorie data", "error", err)
			return nil, fmt.Errorf("failed to fetch calorie data: %w", err)
		}
		calorieData = data
		s.logger.Debug("Fetched calorie data", "count", len(calorieData))
	}

	// Generate Excel file
	fileBytes, err := s.excelGenerator.Generate(weightData, calorieData, req.DataType, req.Language)
	if err != nil {
		s.logger.Error("Failed to generate Excel file", "error", err)
		return nil, fmt.Errorf("failed to generate Excel file: %w", err)
	}

	// Create filename
	fileName := fmt.Sprintf("kkal-export-%s-to-%s.xlsx", req.DateFrom, req.DateTo)

	// Handle delivery
	if req.DeliveryType == DeliveryEmail {
		if req.UserEmail == "" {
			return nil, fmt.Errorf("email address is required for email delivery")
		}

		err = s.emailService.SendEmailWithAttachment(
			req.UserEmail,
			req.Language,
			fileBytes,
			fileName,
		)
		if err != nil {
			s.logger.Error("Failed to send export email", "error", err)
			return nil, fmt.Errorf("failed to send export email: %w", err)
		}

		s.logger.Info("Export sent via email",
			"user_id", req.UserID,
			"email", req.UserEmail,
			"file_name", fileName,
		)

		// Return empty result for email delivery (file was sent)
		return &ExportResult{
			FileName: fileName,
		}, nil
	}

	// Return file for download
	s.logger.Info("Export generated for download",
		"user_id", req.UserID,
		"file_name", fileName,
		"size_bytes", len(fileBytes),
	)

	return &ExportResult{
		FileBytes: fileBytes,
		FileName:  fileName,
	}, nil
}
