package billing

// BillingPreviewRequest represents the query parameters for billing preview
type BillingPreviewRequest struct {
	ClientID  string `form:"client_id" binding:"required"`
	StartDate string `form:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `form:"end_date" binding:"required"`   // YYYY-MM-DD
}

// BillingPreviewResponse represents the calculated billing preview
type BillingPreviewResponse struct {
	ClientID    string             `json:"client_id"`
	Period      PeriodInfo         `json:"period"`
	UsageKwh    float64            `json:"usage_kwh"`
	Calculation CalculationDetails `json:"calculation"`
}

// PeriodInfo represents the billing period
type PeriodInfo struct {
	Start string `json:"start"` // YYYY-MM-DD
	End   string `json:"end"`   // YYYY-MM-DD
}

// CalculationDetails shows the min and max values used in calculation
type CalculationDetails struct {
	MinKwh float64 `json:"min_kwh"`
	MaxKwh float64 `json:"max_kwh"`
}

// UsageData represents the raw data retrieved from the database
type UsageData struct {
	MinKwh *float64
	MaxKwh *float64
}
