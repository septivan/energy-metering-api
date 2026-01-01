package timeseries

import (
	"context"

	"go.uber.org/zap"

	"energy-metering-api/internal/repository"
)

// Service handles timeseries data logic
type Service struct {
	repo   *repository.Repository
	logger *zap.Logger
}

// NewService creates a new timeseries service
func NewService(repo *repository.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetTimeseries retrieves timeseries data for a client
func (s *Service) GetTimeseries(ctx context.Context, clientID string) (*TimeseriesResponse, error) {
	s.logger.Info("fetching timeseries data", zap.String("client_id", clientID))

	// Query database
	data, err := s.repo.GetTimeseriesData(ctx, clientID)
	if err != nil {
		s.logger.Error("failed to get timeseries data", zap.Error(err))
		return nil, err
	}

	// Initialize metrics with all four metric types as empty arrays
	metrics := map[string][]DataPoint{
		"Total_Import_kWh": {},
		"Volts":            {},
		"Current":          {},
		"Active_Power":     {},
	}

	// Metric name mapping from database to response
	metricNameMap := map[string]string{
		"Voltage": "Volts",
		"Power":   "Active_Power",
	}

	// Group data by metric_name
	for _, d := range data {
		// Map the metric name if it needs to be renamed
		metricName := d.MetricName
		if mappedName, exists := metricNameMap[d.MetricName]; exists {
			metricName = mappedName
		}

		metrics[metricName] = append(metrics[metricName], DataPoint{
			Ts:     d.ReadingTimestamp,
			Value:  d.Value,
			Status: d.Status,
		})
	}

	response := &TimeseriesResponse{
		ClientID: clientID,
		Metrics:  metrics,
	}

	s.logger.Info("timeseries data retrieved",
		zap.String("client_id", clientID),
		zap.Int("total_points", len(data)),
		zap.Int("metric_count", len(metrics)),
	)

	return response, nil
}
