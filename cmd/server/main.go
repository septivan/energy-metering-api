package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"energy-metering-api/internal/anomalies"
	"energy-metering-api/internal/billing"
	"energy-metering-api/internal/config"
	"energy-metering-api/internal/dashboard"
	"energy-metering-api/internal/db"
	"energy-metering-api/internal/handler"
	"energy-metering-api/internal/mq"
	"energy-metering-api/internal/pdf"
	"energy-metering-api/internal/repository"
	"energy-metering-api/internal/service"
	"energy-metering-api/internal/timeseries"
	"energy-metering-api/internal/websocket"
)

func NewRouter() *gin.Engine {
	return gin.New()
}

func main() {
	// Load .env file - try multiple locations for flexibility
	loaded := false
	envPaths := []string{
		".env",                    // Current directory (for pods/containers)
		"../../.env",             // From cmd/server to project root
		filepath.Join("..", "..", ".env"), // Cross-platform path
	}
	
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("Loaded .env from: %s\n", path)
			loaded = true
			break
		}
	}
	
	if !loaded {
		fmt.Println("Warning: .env file not found in any location, using environment variables")
	}
	
	app := fx.New(
		fx.Provide(
			// Core dependencies
			config.New,
			newLogger,
			func(cfg *config.Config, lc fx.Lifecycle) (*pgxpool.Pool, error) {
				return db.NewPool(lc, cfg.DatabaseURL)
			},
			
			// Repository layer
			repository.NewRepository,
			
			// Service layer
			service.NewService,
			billing.NewService,
			dashboard.NewService,
			timeseries.NewService,
			anomalies.NewService,
			
			// Handler layer
			handler.NewHandler,
			billing.NewHandler,
			dashboard.NewHandler,
			timeseries.NewHandler,
			anomalies.NewHandler,
			
			// Infrastructure
			websocket.NewHub,
			mq.NewConsumer,
			pdf.NewGenerator,
			
			// HTTP router
			NewRouter,
		),
		fx.Invoke(startServer),
	)

	// Load config for timeouts
	cfg, err := config.New()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Start the application
	startCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("Shutting down gracefully...")

	// Stop the application with timeout
	stopCtx, cancelStop := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelStop()
	if err := app.Stop(stopCtx); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
	}
}

func startServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger *zap.Logger,
	pool *pgxpool.Pool,
	h *handler.Handler,
	billingHandler *billing.Handler,
	dashboardHandler *dashboard.Handler,
	timeseriesHandler *timeseries.Handler,
	anomaliesHandler *anomalies.Handler,
	hub *websocket.Hub,
	consumer *mq.Consumer,
	router *gin.Engine,
) {
	// Register routes
	RegisterRoutes(router, h, billingHandler, dashboardHandler, timeseriesHandler, anomaliesHandler, hub, cfg, logger)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServicePort),
		Handler: router,
	}

	// Create a context for long-running goroutines that will be cancelled on shutdown
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// Start background services
			go hub.Run()
			go consumer.Start(ctx)
			
			// Start HTTP server
			go func() {
				logger.Info("starting http server",
					zap.String("service", cfg.ServiceName),
					zap.Int("port", cfg.ServicePort),
				)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("http server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping services...")
			cancel() // Cancel the long-running context
			
			// Shutdown HTTP server
			if err := srv.Shutdown(ctx); err != nil {
				logger.Error("error shutting down http server", zap.Error(err))
			}
			
			// Close database pool
			pool.Close()
			logger.Info("services stopped successfully")
			return nil
		},
	})
}
