package timeseries

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles timeseries HTTP requests
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new timeseries handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetTimeseries handles GET /api/timeseries
func (h *Handler) GetTimeseries(c *gin.Context) {
	var req TimeseriesRequest

	// Bind and validate query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("invalid request parameters", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid parameters",
			"details": err.Error(),
		})
		return
	}

	// Get timeseries data
	result, err := h.service.GetTimeseries(c.Request.Context(), req.ClientID)
	if err != nil {
		h.logger.Error("failed to get timeseries", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
