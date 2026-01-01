package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"energy-metering-api/internal/config"
	"energy-metering-api/internal/pdf"
	"energy-metering-api/internal/service"
	"energy-metering-api/internal/websocket"
)

type Handler struct {
	svc    *service.Service
	pdfGen *pdf.Generator
	hub    *websocket.Hub
	logger *zap.Logger
	cfg    *config.Config
}

func NewHandler(svc *service.Service, pdfGen *pdf.Generator, hub *websocket.Hub, logger *zap.Logger, cfg *config.Config) *Handler {
	return &Handler{svc: svc, pdfGen: pdfGen, hub: hub, logger: logger, cfg: cfg}
}

// Deprecated: GetLatest endpoint not used anymore
// func (h *Handler) GetLatest(c *gin.Context) {
// 	limit := h.cfg.DefaultLimit
// 	if l := c.Query("limit"); l != "" {
// 		if v, err := strconv.Atoi(l); err == nil {
// 			limit = v
// 		}
// 	}
// 	ctx := c.Request.Context()
// 	res, err := h.svc.LatestReadings(ctx, limit)
// 	if err != nil {
// 		h.logger.Error("GetLatest error", zap.Error(err))
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, res)
// }

// Deprecated: GetTimeSeries endpoint not used anymore - use /api/v1/timeseries instead
// func (h *Handler) GetTimeSeries(c *gin.Context) {
// 	client := c.Query("client_id")
// 	metric := c.Query("metric")
// 	fromS := c.Query("from")
// 	toS := c.Query("to")
// 	if client == "" || metric == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id and metric are required"})
// 		return
// 	}
// 	from := time.Now().Add(-time.Duration(h.cfg.DefaultTimeRangeHours) * time.Hour)
// 	to := time.Now()
// 	if fromS != "" {
// 		if t, err := time.Parse(time.RFC3339, fromS); err == nil {
// 			from = t
// 		}
// 	}
// 	if toS != "" {
// 		if t, err := time.Parse(time.RFC3339, toS); err == nil {
// 			to = t
// 		}
// 	}
// 	pts, err := h.svc.TimeSeries(c.Request.Context(), client, metric, from, to)
// 	if err != nil {
// 		h.logger.Error("GetTimeSeries error", zap.Error(err))
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, pts)
// }

// Deprecated: GetInvoice endpoint not used anymore - use /api/v1/billing or /api/v1/billing/preview instead
// func (h *Handler) GetInvoice(c *gin.Context) {
// 	c.JSON(http.StatusGone, gin.H{
// 		"error": "This endpoint is deprecated. Please use /api/v1/billing/preview with client_id, start_date, and end_date query parameters",
// 		"new_endpoint": "/api/v1/billing/preview",
// 	})
// }

func (h *Handler) GetClients(c *gin.Context) {
	clients, err := h.svc.GetAllClients(c.Request.Context())
	if err != nil {
		h.logger.Error("GetClients error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, clients)
}
