package timeseries

import "time"

// TimeseriesRequest represents query parameters for timeseries API
type TimeseriesRequest struct {
	ClientID string `form:"client_id" binding:"required,uuid"`
}

// TimeseriesResponse represents the timeseries data grouped by metric
type TimeseriesResponse struct {
	ClientID string                 `json:"client_id"`
	Metrics  map[string][]DataPoint `json:"metrics"`
}

// PeriodInfo represents the time range of data
type PeriodInfo struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// DataPoint represents a single reading point
type DataPoint struct {
	Ts     time.Time `json:"ts"`
	Value  float64   `json:"value"`
	Status string    `json:"status"`
}

// ReadingData represents raw data from database
type ReadingData struct {
	MetricName       string
	ReadingTimestamp time.Time
	Value            float64
	Status           string
}
