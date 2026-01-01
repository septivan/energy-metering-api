package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler handles dashboard-related HTTP requests
type Handler struct {
	service *Service
	logger  *zap.Logger
}

// NewHandler creates a new dashboard handler
func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetSummary handles GET /api/dashboard/summary
// Returns dashboard statistics for today and yesterday
func (h *Handler) GetSummary(c *gin.Context) {
	ctx := c.Request.Context()

	summary, err := h.service.GetSummary(ctx)
	if err != nil {
		h.logger.Error("failed to get dashboard summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve dashboard summary",
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}
