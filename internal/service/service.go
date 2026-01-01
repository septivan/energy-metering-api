package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"energy-metering-api/internal/repository"
)

type Service struct {
	repo   *repository.Repository
	logger *zap.Logger
}

func NewService(repo *repository.Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) LatestReadings(ctx context.Context, limit int) ([]repository.Reading, error) {
	return s.repo.GetLatestReadings(ctx, limit)
}

func (s *Service) TimeSeries(ctx context.Context, clientID, metric string, from, to time.Time) ([]repository.TimeSeriesPoint, error) {
	return s.repo.GetTimeSeries(ctx, clientID, metric, from, to)
}

func (s *Service) GetAllClients(ctx context.Context) ([]repository.ClientInfo, error) {
	return s.repo.GetAllClients(ctx)
}
