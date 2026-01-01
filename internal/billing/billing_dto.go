package billing

import "time"

// BillingRequest represents query parameters for billing calculation
type BillingRequest struct {
	ClientID  string `form:"client_id" binding:"required"`
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
}

// BillingResponse represents the comprehensive billing calculation result
type BillingResponse struct {
	ClientID         string            `json:"client_id"`
	Period           BillingPeriod     `json:"period"`
	SectionA         TotalKwhSection   `json:"section_a_total_kwh"`
	SectionB         RecalculatedSection `json:"section_b_recalculated_kwh"`
	SectionC         SummarySection    `json:"section_c_summary"`
}

// BillingPeriod represents the billing period
type BillingPeriod struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// TotalKwhSection represents baseline calculation (MAX - MIN)
type TotalKwhSection struct {
	MinKwh   float64 `json:"min_kwh"`
	MaxKwh   float64 `json:"max_kwh"`
	TotalKwh float64 `json:"total_kwh"`
}

// RecalculatedSection represents formula-based calculation
type RecalculatedSection struct {
	Formula      string  `json:"formula"`
	Voltage      float64 `json:"voltage_v"`
	IntervalSec  int     `json:"interval_seconds"`
	TotalKwh     float64 `json:"total_kwh"`
	Description  string  `json:"description"`
}

// SummarySection represents summary statistics
type SummarySection struct {
	TotalConsumptionKwh float64 `json:"total_consumption_kwh"`
	PeakPowerW          float64 `json:"peak_power_w"`
	MinPowerW           float64 `json:"min_power_w"`
	TotalValidReadings  int     `json:"total_valid_readings"`
	TotalAnomalies      int     `json:"total_anomalies"`
}

// BillingData holds all raw data from database for billing calculation
type BillingData struct {
	// Section A data
	MinTotalImportKwh *float64
	MaxTotalImportKwh *float64
	
	// Section B data - Current readings for recalculation
	CurrentReadings []CurrentReading
	
	// Section C data
	PeakActivePower *float64
	MinActivePower  *float64
	ValidCount      int
	AnomalyCount    int
}

// CurrentReading represents a single current reading for recalculation
type CurrentReading struct {
	Current       float64
	ReadingTime   time.Time
}
