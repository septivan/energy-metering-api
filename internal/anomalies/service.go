package anomalies

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"energy-metering-api/internal/repository"
)

// Service handles anomalies business logic
type Service struct {
	repo   *repository.Repository
	logger *zap.Logger
}

// NewService creates a new anomalies service
func NewService(repo *repository.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetAnomalies retrieves anomaly records with pagination
func (s *Service) GetAnomalies(ctx context.Context, startDate, endDate *time.Time, clientID *string, page, limit int) (*AnomalyResponse, error) {
	s.logger.Info("fetching anomalies",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	// Calculate offset
	offset := (page - 1) * limit

	// Query repository
	records, total, err := s.repo.GetAnomalies(ctx, startDate, endDate, clientID, limit, offset)
	if err != nil {
		s.logger.Error("failed to get anomalies", zap.Error(err))
		return nil, err
	}

	// Convert repository records to response records
	var data []AnomalyRecord
	for _, r := range records {
		data = append(data, AnomalyRecord{
			ID:               r.ID,
			ClientID:         r.ClientID,
			MetricName:       r.MetricName,
			MetricValue:      r.MetricValue,
			ReadingTimestamp: r.ReadingTimestamp,
			ReceivedAt:       r.ReceivedAt,
			ValidationStatus: r.ValidationStatus,
			AnomalyReason:    r.AnomalyReason,
		})
	}

	// Calculate total pages
	totalPage := int(math.Ceil(float64(total) / float64(limit)))

	response := &AnomalyResponse{
		Data: data,
		Pagination: PaginationInfo{
			Page:      page,
			Limit:     limit,
			Total:     total,
			TotalPage: totalPage,
		},
	}

	s.logger.Info("anomalies retrieved",
		zap.Int("total", total),
		zap.Int("returned", len(data)),
	)

	return response, nil
}
