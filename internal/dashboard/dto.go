package dashboard

// SummaryResponse represents the dashboard summary data
type SummaryResponse struct {
	ActiveClientsToday     int                `json:"active_clients_today"`
	ActiveClientsYesterday int                `json:"active_clients_yesterday"`
	ReadingsToday          int                `json:"readings_today"`
	ReadingsYesterday      int                `json:"readings_yesterday"`
	ValidationToday        ValidationBreakdown `json:"validation_today"`
}

// ValidationBreakdown shows counts by validation status
type ValidationBreakdown struct {
	Valid   int `json:"valid"`
	Anomaly int `json:"anomaly"`
	Invalid int `json:"invalid"`
}
