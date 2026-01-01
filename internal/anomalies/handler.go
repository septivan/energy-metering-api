package anomalies

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles anomalies HTTP requests
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new anomalies handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetAnomalies handles GET /api/anomalies
func (h *Handler) GetAnomalies(c *gin.Context) {
	var req AnomalyRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid parameters",
			"details": err.Error(),
		})
		return
	}

	// Parse optional dates (date-only format: YYYY-MM-DD)
	var startDate, endDate *time.Time
	
	if req.From != "" {
		parsed, err := time.Parse("2006-01-02", req.From)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid from format, expected YYYY-MM-DD",
			})
			return
		}
		// Truncate to start of day
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
		startDate = &parsed
	}

	if req.To != "" {
		parsed, err := time.Parse("2006-01-02", req.To)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid to format, expected YYYY-MM-DD",
			})
			return
		}
		// Truncate to start of day and add one day to include the entire end day
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
		parsed = parsed.Add(24 * time.Hour)
		endDate = &parsed
	}

	// Validate date range if both provided
	if startDate != nil && endDate != nil && !startDate.Before(*endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from must be before to",
		})
		return
	}

	// Set default pagination
	page := req.Page
	if page < 1 {
		page = 1
	}

	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	// Prepare client_id pointer
	var clientID *string
	if req.ClientID != "" {
		clientID = &req.ClientID
	}

	// Get anomalies
	result, err := h.service.GetAnomalies(c.Request.Context(), startDate, endDate, clientID, page, limit)
	if err != nil {
		h.logger.Error("failed to get anomalies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
