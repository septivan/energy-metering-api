package anomalies

import "time"

// AnomalyRequest represents query parameters for anomalies API
type AnomalyRequest struct {
	From     string `form:"from"`      // Optional: YYYY-MM-DD
	To       string `form:"to"`        // Optional: YYYY-MM-DD
	ClientID string `form:"client_id"` // Optional UUID
	Page     int    `form:"page"`      // Default 1
	Limit    int    `form:"limit"`     // Default 20
}

// AnomalyResponse represents the anomalies list response
type AnomalyResponse struct {
	Data       []AnomalyRecord `json:"data"`
	Pagination PaginationInfo  `json:"pagination"`
}

// AnomalyRecord represents a single anomaly record
type AnomalyRecord struct {
	ID               string    `json:"id"`
	ClientID         string    `json:"client_id"`
	MetricName       string    `json:"metric_name"`
	MetricValue      float64   `json:"metric_value"`
	ReadingTimestamp time.Time `json:"reading_timestamp"`
	ReceivedAt       time.Time `json:"received_at"`
	ValidationStatus string    `json:"validation_status"`
	AnomalyReason    *string   `json:"anomaly_reason"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}
