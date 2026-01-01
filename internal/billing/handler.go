package billing

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"go.uber.org/zap"

	"energy-metering-api/internal/constants"
)

// Handler handles billing-related HTTP requests
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new billing handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetPreview handles GET /api/v1/billing/preview
// Calculates billing on-demand from raw meter readings
func (h *Handler) GetPreview(c *gin.Context) {
	var req BillingPreviewRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   constants.ErrMissingParameters,
			"details": err.Error(),
		})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.logger.Warn("invalid start_date format", zap.String("start_date", req.StartDate))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid start_date format, expected YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.logger.Warn("invalid end_date format", zap.String("end_date", req.EndDate))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": constants.ErrInvalidDateFormat,
		})
		return
	}

	// Add one day to endDate to include the entire end day
	endDate = endDate.Add(24 * time.Hour)

	// Calculate usage
	result, err := h.service.CalculateUsage(c.Request.Context(), req.ClientID, startDate, endDate)
	if err != nil {
		if errors.Is(err, ErrNoData) {
			h.logger.Info("no data found for billing preview",
				zap.String("client_id", req.ClientID),
				zap.String("start_date", req.StartDate),
				zap.String("end_date", req.EndDate),
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no meter readings found for the specified period",
			})
			return
		}

		if errors.Is(err, ErrInvalidDateRange) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "start_date must be before end_date",
			})
			return
		}

		if errors.Is(err, ErrInvalidUsage) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "data inconsistency detected - max usage is less than min usage",
			})
			return
		}

		h.logger.Error("failed to calculate usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": constants.ErrInternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Deprecated: GetBilling - use GetBillingComprehensive instead
// func (h *Handler) GetBilling(c *gin.Context) {
// 	var req BillingRequest
//
// 	// Bind query parameters
// 	if err := c.ShouldBindQuery(&req); err != nil {
// 		h.logger.Warn("invalid request parameters", zap.Error(err))
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "client_id is required",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	// Parse optional dates
// 	var startDate, endDate *time.Time
//
// 	if req.StartDate != "" {
// 		parsed, err := time.Parse("2006-01-02", req.StartDate)
// 		if err != nil {
// 			h.logger.Warn("invalid start_date format", zap.String("start_date", req.StartDate))
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "invalid start_date format, expected YYYY-MM-DD",
// 			})
// 			return
// 		}
// 		startDate = &parsed
// 	}

// 	if req.EndDate != "" {
// 		parsed, err := time.Parse("2006-01-02", req.EndDate)
// 		if err != nil {
// 			h.logger.Warn("invalid end_date format", zap.String("end_date", req.EndDate))
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "invalid end_date format, expected YYYY-MM-DD",
// 			})
// 			return
// 		}
// 		// Add one day to include the entire end day
// 		parsed = parsed.Add(24 * time.Hour)
// 		endDate = &parsed
// 	}

// 	// Calculate billing
// 	result, err := h.service.CalculateBilling(c.Request.Context(), req.ClientID, startDate, endDate)
// 	if err != nil {
// 		if errors.Is(err, ErrNoData) {
// 			h.logger.Info("no valid data found for billing",
// 				zap.String("client_id", req.ClientID),
// 			)
// 			c.JSON(http.StatusNotFound, gin.H{
// 				"error": "no valid meter readings found for the specified period",
// 			})
// 			return
// 		}

// 		if errors.Is(err, ErrInvalidDateRange) {
// 			c.JSON(http.StatusBadRequest, gin.H{
// 				"error": "start_date must be before end_date",
// 			})
// 			return
// 		}

// 		if errors.Is(err, ErrInvalidUsage) {
// 			c.JSON(http.StatusUnprocessableEntity, gin.H{
// 				"error": "data inconsistency detected - max usage is less than min usage",
// 			})
// 			return
// 		}

// 		h.logger.Error("failed to calculate billing", zap.Error(err))
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": "internal server error",
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, result)
// }

// GetBillingComprehensive handles GET /api/v1/billing - comprehensive billing with all sections
func (h *Handler) GetBillingComprehensive(c *gin.Context) {
	var req BillingRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "client_id, start_date, and end_date are required",
			"details": err.Error(),
		})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		h.logger.Warn("invalid start_date format", zap.String("start_date", req.StartDate))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid start_date format, expected YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		h.logger.Warn("invalid end_date format", zap.String("end_date", req.EndDate))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid end_date format, expected YYYY-MM-DD",
		})
		return
	}

	// Calculate comprehensive billing
	result, err := h.service.CalculateComprehensiveBilling(c.Request.Context(), req.ClientID, startDate, endDate)
	if err != nil {
		if errors.Is(err, ErrNoData) {
			h.logger.Info("no valid data found for billing",
				zap.String("client_id", req.ClientID),
			)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "no valid meter readings found for the specified period",
			})
			return
		}

		if errors.Is(err, ErrInvalidDateRange) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "start_date must be before end_date",
			})
			return
		}

		if errors.Is(err, ErrInvalidUsage) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "data inconsistency detected - max usage is less than min usage",
			})
			return
		}

		h.logger.Error("failed to calculate billing", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Check if PDF is requested
	if c.Query("format") == "pdf" {
		pdfBytes, err := h.generateBillingPDF(result)
		if err != nil {
			h.logger.Error("failed to generate PDF", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to generate PDF",
			})
			return
		}

		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=billing_%s_%s.pdf", req.ClientID[:8], req.StartDate))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
		return
	}

	// Return JSON by default
	c.JSON(http.StatusOK, result)
}

// generateBillingPDF generates a PDF from billing data
func (h *Handler) generateBillingPDF(data *BillingResponse) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 18)

	// === HEADER ===
	pdf.Cell(0, 10, "ENERGY METERING BILLING REPORT")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, fmt.Sprintf("Client ID: %s", data.ClientID))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Period: %s to %s", data.Period.From, data.Period.To))
	pdf.Ln(12)

	// === SECTION A ===
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "SECTION A: Total kWh (Baseline)")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Minimum Reading: %.2f kWh", data.SectionA.MinKwh))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Maximum Reading: %.2f kWh", data.SectionA.MaxKwh))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Total Consumption: %.2f kWh", data.SectionA.TotalKwh))
	pdf.Ln(12)

	// === SECTION B ===
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "SECTION B: Recalculated Total kWh (Formula-Based)")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Formula: %s", data.SectionB.Formula))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Total Recalculated: %.4f kWh", data.SectionB.TotalKwh))
	pdf.Ln(12)

	// === SECTION C ===
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "SECTION C: Summary Statistics")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Total Consumption: %.2f kWh", data.SectionC.TotalConsumptionKwh))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Peak Power: %.2f W", data.SectionC.PeakPowerW))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Minimum Power: %.2f W", data.SectionC.MinPowerW))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Valid Readings: %d", data.SectionC.TotalValidReadings))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Anomaly Readings: %d", data.SectionC.TotalAnomalies))
	pdf.Ln(15)

	// === FOOTER ===
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(0, 5, "This is an informational billing report only. Not an invoice.")

	// Check for errors
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("PDF generation error: %w", err)
	}

	// Output to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to output PDF: %w", err)
	}

	return buf.Bytes(), nil
}
