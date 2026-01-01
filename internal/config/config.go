package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName  string
	ServicePort  int
	DatabaseURL  string
	RabbitMQURL  string
	ReadOnlyDB   bool
	
	// API Configuration
	APIBasePath string
	WSPath      string
	ShutdownTimeout time.Duration
	
	// Handler Configuration
	DefaultLimit         int
	DefaultTimeRangeHours int
	
	// RabbitMQ Configuration
	RabbitMQExchange         string
	RabbitMQQueue            string
	RabbitMQRoutingKey       string
	RabbitMQRetryInterval    time.Duration
	
	// PDF Configuration
	PDFOutputDir string
	
	// WebSocket Configuration
	WSBufferSize       int
	WSClientBufferSize int
}

func New() (*Config, error) {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "energy-metering-api"
	}

	port := 8080
	if p := os.Getenv("SERVICE_PORT"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid SERVICE_PORT: %w", err)
		}
		port = v
	}
	
	apiBasePath := os.Getenv("API_BASE_PATH")
	if apiBasePath == "" {
		apiBasePath = "/api/v1"
	}
	// Prepend service name to API base path
	apiBasePath = "/" + serviceName + apiBasePath
	
	wsPath := os.Getenv("WS_PATH")
	if wsPath == "" {
		wsPath = "/ws/live"
	}
	// Prepend service name to WebSocket path
	wsPath = "/" + serviceName + wsPath
	
	shutdownTimeout := 15 * time.Second
	if s := os.Getenv("SHUTDOWN_TIMEOUT"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			shutdownTimeout = time.Duration(v) * time.Second
		}
	}
	
	defaultLimit := 50
	if l := os.Getenv("DEFAULT_LIMIT"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			defaultLimit = v
		}
	}
	
	defaultTimeRangeHours := 24
	if t := os.Getenv("DEFAULT_TIME_RANGE_HOURS"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			defaultTimeRangeHours = v
		}
	}
	
	rabbitMQExchange := os.Getenv("RABBITMQ_EXCHANGE")
	if rabbitMQExchange == "" {
		rabbitMQExchange = "energy-metering.worker.events.exchange"
	}
	
	rabbitMQQueue := os.Getenv("RABBITMQ_QUEUE")
	if rabbitMQQueue == "" {
		rabbitMQQueue = "energy-metering.ws.bridge"
	}
	
	rabbitMQRoutingKey := os.Getenv("RABBITMQ_ROUTING_KEY")
	if rabbitMQRoutingKey == "" {
		rabbitMQRoutingKey = "#"
	}
	
	rabbitMQRetryInterval := 5 * time.Second
	if r := os.Getenv("RABBITMQ_RETRY_INTERVAL"); r != "" {
		if v, err := strconv.Atoi(r); err == nil {
			rabbitMQRetryInterval = time.Duration(v) * time.Second
		}
	}
	
	pdfOutputDir := os.Getenv("PDF_OUTPUT_DIR")
	if pdfOutputDir == "" {
		pdfOutputDir = "./tmp"
	}
	
	wsBufferSize := 256
	if w := os.Getenv("WS_BUFFER_SIZE"); w != "" {
		if v, err := strconv.Atoi(w); err == nil {
			wsBufferSize = v
		}
	}
	
	wsClientBufferSize := 256
	if w := os.Getenv("WS_CLIENT_BUFFER_SIZE"); w != "" {
		if v, err := strconv.Atoi(w); err == nil {
			wsClientBufferSize = v
		}
	}

	cfg := &Config{
		ServiceName:  serviceName,
		ServicePort:  port,
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RabbitMQURL:  os.Getenv("RABBITMQ_URL"),
		ReadOnlyDB:   true,
		APIBasePath:  apiBasePath,
		WSPath:       wsPath,
		ShutdownTimeout: shutdownTimeout,
		DefaultLimit: defaultLimit,
		DefaultTimeRangeHours: defaultTimeRangeHours,
		RabbitMQExchange: rabbitMQExchange,
		RabbitMQQueue: rabbitMQQueue,
		RabbitMQRoutingKey: rabbitMQRoutingKey,
		RabbitMQRetryInterval: rabbitMQRetryInterval,
		PDFOutputDir: pdfOutputDir,
		WSBufferSize: wsBufferSize,
		WSClientBufferSize: wsClientBufferSize,
	}
	return cfg, nil
}
