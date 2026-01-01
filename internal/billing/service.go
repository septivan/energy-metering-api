package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"energy-metering-api/internal/constants"
	"energy-metering-api/internal/repository"
)

var (
	ErrNoData          = errors.New("no meter readings found for the specified period")
	ErrInvalidDateRange = errors.New("start_date must be before end_date")
	ErrInvalidUsage    = errors.New("max usage is less than min usage - data inconsistency")
)

// Service handles billing calculation logic
type Service struct {
	repo   *repository.Repository
	logger *zap.Logger
}

// NewService creates a new billing service
func NewService(repo *repository.Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CalculateUsage calculates billing preview from raw meter readings
func (s *Service) CalculateUsage(ctx context.Context, clientID string, startDate, endDate time.Time) (*BillingPreviewResponse, error) {
	// Validate date range (allow equal dates for single-day billing)
	if startDate.After(endDate) {
		return nil, ErrInvalidDateRange
	}

	s.logger.Info("calculating usage",
		zap.String("client_id", clientID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	// Query raw meter readings
	data, err := s.repo.GetUsageData(ctx, clientID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get usage data", zap.Error(err))
		return nil, fmt.Errorf("failed to query meter readings: %w", err)
	}

	// Check if data exists
	if data.MinKwh == nil || data.MaxKwh == nil {
		s.logger.Warn("no data found for billing calculation",
			zap.String("client_id", clientID),
			zap.Time("start_date", startDate),
			zap.Time("end_date", endDate),
		)
		return nil, ErrNoData
	}

	minKwh := *data.MinKwh
	maxKwh := *data.MaxKwh

	// Validate data consistency
	if maxKwh < minKwh {
		s.logger.Error("data inconsistency detected",
			zap.String("client_id", clientID),
			zap.Float64("min_kwh", minKwh),
			zap.Float64("max_kwh", maxKwh),
		)
		return nil, ErrInvalidUsage
	}

	// Calculate usage
	usageKwh := maxKwh - minKwh

	s.logger.Info("usage calculated successfully",
		zap.String("client_id", clientID),
		zap.Float64("usage_kwh", usageKwh),
		zap.Float64("min_kwh", minKwh),
		zap.Float64("max_kwh", maxKwh),
	)

	return &BillingPreviewResponse{
		ClientID: clientID,
		Period: PeriodInfo{
			Start: startDate.Format(constants.DateFormatYMD),
			End:   endDate.Format(constants.DateFormatYMD),
		},
		UsageKwh: usageKwh,
		Calculation: CalculationDetails{
			MinKwh: minKwh,
			MaxKwh: maxKwh,
		},
	}, nil
}

// CalculateBilling calculates billing for a client within a date range
// Deprecated: CalculateBilling - use CalculateComprehensiveBilling instead
// func (s *Service) CalculateBilling(ctx context.Context, clientID string, startDate, endDate *time.Time) (*BillingResponse, error) {
// 	// Set default dates if not provided
// 	now := time.Now()
// 	var start, end time.Time
// 	
// 	if startDate == nil {
// 		// Default: start of current month
// 		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
// 	} else {
// 		start = *startDate
// 	}
// 	
// 	if endDate == nil {
// 		// Default: end of current month
// 		end = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
// 	} else {
// 		end = *endDate
// 	}
// 	
// 	// Validate date range
// 	if !start.Before(end) {
// 		return nil, ErrInvalidDateRange
// 	}

// 	s.logger.Info("calculating billing",
// 		zap.String("client_id", clientID),
// 		zap.Time("start_date", start),
// 		zap.Time("end_date", end),
// 	)

// 	// Query raw meter readings with VALID status
// 	data, err := s.repo.GetUsageData(ctx, clientID, start, end)
// 	if err != nil {
// 		s.logger.Error("failed to get usage data", zap.Error(err))
// 		return nil, fmt.Errorf("failed to query meter readings: %w", err)
// 	}

// 	// Check if data exists
// 	if data.MinKwh == nil || data.MaxKwh == nil {
// 		s.logger.Warn("no valid data found for billing calculation",
// 			zap.String("client_id", clientID),
// 			zap.Time("start_date", start),
// 			zap.Time("end_date", end),
// 		)
// 		return nil, ErrNoData
// 	}

// 	minKwh := *data.MinKwh
// 	maxKwh := *data.MaxKwh

// 	// Validate data consistency
// 	if maxKwh < minKwh {
// 		s.logger.Error("data inconsistency detected",
// 			zap.String("client_id", clientID),
// 			zap.Float64("min_kwh", minKwh),
// 			zap.Float64("max_kwh", maxKwh),
// 		)
// 		return nil, ErrInvalidUsage
// 	}

// 	// Calculate billing
// 	billingKwh := maxKwh - minKwh

// 	s.logger.Info("billing calculated successfully",
// 		zap.String("client_id", clientID),
// 		zap.Float64("billing_kwh", billingKwh),
// 	)

// 	response := BillingResponse{
// 		ClientID: clientID,
// 		Period: BillingPeriod{
// 			From: start.Format("2006-01-02"),
// 			To:   end.Format("2006-01-02"),
// 		},
// 		TotalKwh: billingKwh,
// 	}

// 	return &response, nil
// }


// CalculateComprehensiveBilling calculates comprehensive billing with all sections
func (s *Service) CalculateComprehensiveBilling(ctx context.Context, clientID string, startDate, endDate time.Time) (*BillingResponse, error) {
	// Validate date range (allow equal dates for single-day billing)
	if startDate.After(endDate) {
		return nil, ErrInvalidDateRange
	}

	s.logger.Info("calculating comprehensive billing",
		zap.String("client_id", clientID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	// Get all billing data
	data, currentReadings, err := s.repo.GetComprehensiveBillingData(ctx, clientID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get billing data", zap.Error(err))
		return nil, fmt.Errorf("failed to query billing data: %w", err)
	}

	// Check if data exists
	if data.MinTotalImportKwh == nil || data.MaxTotalImportKwh == nil {
		s.logger.Warn("no valid data found for billing calculation",
			zap.String("client_id", clientID),
		)
		return nil, ErrNoData
	}

	// === SECTION A: Total kWh (Baseline) ===
	minKwh := *data.MinTotalImportKwh
	maxKwh := *data.MaxTotalImportKwh
	
	if maxKwh < minKwh {
		s.logger.Error("data inconsistency detected",
			zap.String("client_id", clientID),
			zap.Float64("min_kwh", minKwh),
			zap.Float64("max_kwh", maxKwh),
		)
		return nil, ErrInvalidUsage
	}
	
	totalKwhBaseline := maxKwh - minKwh

	// === SECTION B: Recalculated kWh (Formula) ===
	var recalculatedKwh float64
	for _, reading := range currentReadings {
		// kWh_per_row = (Voltage × Current × Time(h)) / 1000
		kwhPerRow := (constants.StandardVoltage * reading.Current * constants.ReadingIntervalHours) / 1000.0
		recalculatedKwh += kwhPerRow
	}

	s.logger.Info("section B calculation",
		zap.Int("current_readings", len(currentReadings)),
		zap.Float64("recalculated_kwh", recalculatedKwh),
	)

	// === SECTION C: Summary Statistics ===
	peakPowerW := 0.0
	minPowerW := 0.0
	if data.PeakActivePower != nil {
		peakPowerW = *data.PeakActivePower
	}
	if data.MinActivePower != nil {
		minPowerW = *data.MinActivePower
	}

	s.logger.Info("comprehensive billing calculated successfully",
		zap.String("client_id", clientID),
		zap.Float64("baseline_kwh", totalKwhBaseline),
		zap.Float64("recalculated_kwh", recalculatedKwh),
		zap.Float64("peak_power_w", peakPowerW),
		zap.Int("valid_readings", data.ValidCount),
		zap.Int("anomalies", data.AnomalyCount),
	)

	response := BillingResponse{
		ClientID: clientID,
		Period: BillingPeriod{
			From: startDate.Format(constants.DateFormatYMD),
			To:   endDate.Format(constants.DateFormatYMD),
		},
		SectionA: TotalKwhSection{
			MinKwh:   minKwh,
			MaxKwh:   maxKwh,
			TotalKwh: totalKwhBaseline,
		},
		SectionB: RecalculatedSection{
			Formula:      constants.FormulaKwh,
			Voltage:      constants.StandardVoltage,
			IntervalSec:  constants.ReadingIntervalSec,
			TotalKwh:     recalculatedKwh,
			Description:  "Calculated from VALID Current readings using 230V and 10s interval",
		},
		SectionC: SummarySection{
			TotalConsumptionKwh: totalKwhBaseline, // Use baseline as primary consumption
			PeakPowerW:          peakPowerW,
			MinPowerW:           minPowerW,
			TotalValidReadings:  data.ValidCount,
			TotalAnomalies:      data.AnomalyCount,
		},
	}

	return &response, nil
}
