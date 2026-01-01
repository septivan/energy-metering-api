package constants

// HTTP Status Messages
const (
	ErrInternalServer      = "internal server error"
	ErrBadRequest          = "bad request"
	ErrNotFound            = "not found"
	ErrInvalidDateFormat   = "invalid date format, expected YYYY-MM-DD"
	ErrMissingParameters   = "missing required parameters"
	ErrInvalidParameters   = "invalid parameters"
)

// Validation Status
const (
	StatusValid   = "valid"
	StatusInvalid = "invalid"
	StatusAnomaly = "anomaly"
	StatusVALID   = "VALID" // Uppercase variant used in some queries
)

// Metric Names
const (
	MetricTotalImportKwh = "Total_Import_kWh"
	MetricVolts          = "Volts"
	MetricCurrent        = "Current"
	MetricActivePower    = "Active_Power"
)

// Date Formats
const (
	DateFormatYMD      = "2006-01-02"
	DateFormatYMDHMS   = "2006-01-02 15:04:05"
	DateFormatRFC3339  = "2006-01-02T15:04:05Z07:00"
)

// Calculation Constants
const (
	StandardVoltage      = 230.0  // Volts
	ReadingIntervalSec   = 10     // seconds
	ReadingIntervalHours = float64(ReadingIntervalSec) / 3600.0
)

// Formula descriptions
const (
	FormulaKwh = "kWh = (Voltage × Current × Time) / 1000"
)
