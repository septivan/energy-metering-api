package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"energy-metering-api/internal/anomalies"
	"energy-metering-api/internal/billing"
	"energy-metering-api/internal/config"
	"energy-metering-api/internal/dashboard"
	"energy-metering-api/internal/handler"
	"energy-metering-api/internal/timeseries"
	"energy-metering-api/internal/websocket"
)

// RegisterRoutes registers HTTP routes on the provided Gin engine
func RegisterRoutes(
	r *gin.Engine,
	h *handler.Handler,
	billingHandler *billing.Handler,
	dashboardHandler *dashboard.Handler,
	timeseriesHandler *timeseries.Handler,
	anomaliesHandler *anomalies.Handler,
	hub *websocket.Hub,
	cfg *config.Config,
	logger *zap.Logger,
) {
	// Recovery middleware
	r.Use(gin.Recovery())
	
	// Logging middleware
	r.Use(loggingMiddleware(logger))
	
	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": cfg.ServiceName})
	})

	// API routes
	api := r.Group(cfg.APIBasePath)
	{
		// Billing endpoints
		billingGroup := api.Group("/billing")
		{
			billingGroup.GET("", billingHandler.GetBillingComprehensive)
			billingGroup.GET("/preview", billingHandler.GetPreview)
		}
		
		// Dashboard endpoints
		dashboardGroup := api.Group("/dashboard")
		{
			dashboardGroup.GET("/summary", dashboardHandler.GetSummary)
		}
		
		// Timeseries endpoints
		api.GET("/timeseries", timeseriesHandler.GetTimeseries)
		
		// Client endpoints
		api.GET("/clients", h.GetClients)
		
		// Anomalies endpoints
		api.GET("/anomalies", anomaliesHandler.GetAnomalies)
	}

	// WebSocket endpoint
	r.GET(cfg.WSPath, func(c *gin.Context) {
		websocket.ServeWS(hub, c.Writer, c.Request)
	})
}

// loggingMiddleware creates a middleware for logging HTTP requests
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after processing
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		logger.Info("http request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
		)
	}
}
