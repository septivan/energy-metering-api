package dashboard

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"energy-metering-api/internal/repository"
)

// Service handles dashboard business logic
type Service struct {
	repo   *repository.Repository
	logger *zap.Logger
}

// NewService creates a new dashboard service
func NewService(repo *repository.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetSummary retrieves dashboard summary statistics
func (s *Service) GetSummary(ctx context.Context) (*SummaryResponse, error) {
	s.logger.Info("fetching dashboard summary")

	// Get stats from repository
	stats, err := s.repo.GetDashboardStats(ctx)
	if err != nil {
		s.logger.Error("failed to get dashboard stats", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve dashboard statistics: %w", err)
	}

	// Map to response DTO
	response := &SummaryResponse{
		ActiveClientsToday:     stats.ActiveClientsToday,
		ActiveClientsYesterday: stats.ActiveClientsYesterday,
		ReadingsToday:          stats.ReadingsToday,
		ReadingsYesterday:      stats.ReadingsYesterday,
		ValidationToday: ValidationBreakdown{
			Valid:   stats.ValidToday,
			Anomaly: stats.AnomalyToday,
			Invalid: stats.InvalidToday,
		},
	}

	s.logger.Info("dashboard summary retrieved successfully",
		zap.Int("active_clients_today", stats.ActiveClientsToday),
		zap.Int("readings_today", stats.ReadingsToday),
	)

	return response, nil
}
