package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Reading struct {
	ClientID string    `json:"client_id"`
	Metric   string    `json:"metric_name"`
	Value    float64   `json:"metric_value"`
	Ts       time.Time `json:"reading_timestamp"`
}

type TimeSeriesPoint struct {
	Ts    time.Time `json:"ts"`
	Value float64   `json:"value"`
}

type Repository struct {
	pool *pgxpool.Pool
} // Read-only access to meter data

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetLatestReadings(ctx context.Context, limit int) ([]Reading, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT client_id::text, metric_name, metric_value, reading_timestamp
        FROM meter_readings_raw
        ORDER BY reading_timestamp DESC
        LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Reading
	for rows.Next() {
		var rd Reading
		if err := rows.Scan(&rd.ClientID, &rd.Metric, &rd.Value, &rd.Ts); err != nil {
			return nil, err
		}
		out = append(out, rd)
	}
	return out, nil
}

func (r *Repository) GetTimeSeries(ctx context.Context, clientID, metric string, from, to time.Time) ([]TimeSeriesPoint, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT reading_timestamp, metric_value
        FROM meter_readings_raw
        WHERE client_id::text = $1 AND metric_name = $2 AND reading_timestamp >= $3 AND reading_timestamp <= $4
        ORDER BY reading_timestamp ASC
    `, clientID, metric, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pts []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Ts, &p.Value); err != nil {
			return nil, err
		}
		pts = append(pts, p)
	}
	return pts, nil
}

// GetUsageData retrieves min and max Total_Import_kWh for a client within a date range
func (r *Repository) GetUsageData(ctx context.Context, clientID string, startDate, endDate time.Time) (*UsageData, error) {
	query := `
		SELECT
			MIN(metric_value) AS min_kwh,
			MAX(metric_value) AS max_kwh
		FROM meter_readings_raw
		WHERE client_id = $1
			AND metric_name = 'Total_Import_kWh'
			AND validation_status = 'VALID'
			AND reading_timestamp >= $2
			AND reading_timestamp < $3
	`

	// Log query and parameters
	// fmt.Println("=== GET BILLING QUERY ===")
	// fmt.Println(query)
	// fmt.Println("=== PARAMETERS ===")
	// fmt.Printf("$1 (client_id): %s\n", clientID)
	// fmt.Printf("$2 (start_date): %s\n", startDate.Format("2006-01-02 15:04:05"))
	// fmt.Printf("$3 (end_date): %s\n", endDate.Format("2006-01-02 15:04:05"))
	// fmt.Println("======================")

	var data UsageData
	err := r.pool.QueryRow(ctx, query, clientID, startDate, endDate).Scan(&data.MinKwh, &data.MaxKwh)
	if err != nil {
		// fmt.Println("=== QUERY ERROR ===")
		// fmt.Printf("Error: %v\n", err)
		// fmt.Println("===================")
		return nil, err
	}

	// Log output
	// fmt.Println("=== QUERY RESULT ===")
	// if data.MinKwh != nil {
	// 	fmt.Printf("min_kwh: %f\n", *data.MinKwh)
	// } else {
	// 	fmt.Println("min_kwh: NULL")
	// }
	// if data.MaxKwh != nil {
	// 	fmt.Printf("max_kwh: %f\n", *data.MaxKwh)
	// } else {
	// 	fmt.Println("max_kwh: NULL")
	// }
	// fmt.Println("====================")

	return &data, nil
}

// GetComprehensiveBillingData retrieves all billing data in one efficient query
func (r *Repository) GetComprehensiveBillingData(ctx context.Context, clientID string, startDate, endDate time.Time) (*BillingData, []CurrentReading, error) {
	// Section A & C: Get aggregated data
	aggregateQuery := `
		SELECT
			-- Section A: Total Import kWh
			MIN(CASE WHEN metric_name = 'Total_Import_kWh' AND validation_status = 'valid' THEN metric_value END) AS min_total_import_kwh,
			MAX(CASE WHEN metric_name = 'Total_Import_kWh' AND validation_status = 'valid' THEN metric_value END) AS max_total_import_kwh,
			-- Section C: Power statistics
			MAX(CASE WHEN metric_name = 'Active_Power' AND validation_status = 'valid' THEN metric_value END) AS peak_active_power,
			MIN(CASE WHEN metric_name = 'Active_Power' AND validation_status = 'valid' THEN metric_value END) AS min_active_power,
			-- Section C: Counts
			COUNT(CASE WHEN validation_status = 'valid' THEN 1 END) AS valid_count,
			COUNT(CASE WHEN validation_status IN ('anomaly', 'invalid') THEN 1 END) AS anomaly_count
		FROM meter_readings_raw
		WHERE client_id = $1
			AND date_trunc('day', reading_timestamp) >= $2::date
			AND date_trunc('day', reading_timestamp) <= $3::date
	`

	// fmt.Println("=== COMPREHENSIVE BILLING AGGREGATE QUERY ===")
	// fmt.Println(aggregateQuery)
	// fmt.Printf("$1 (client_id): %s\n", clientID)
	// fmt.Printf("$2 (start_date): %s\n", startDate.Format("2006-01-02"))
	// fmt.Printf("$3 (end_date): %s\n", endDate.Format("2006-01-02"))

	var data BillingData
	err := r.pool.QueryRow(ctx, aggregateQuery, clientID, startDate, endDate).Scan(
		&data.MinTotalImportKwh,
		&data.MaxTotalImportKwh,
		&data.PeakActivePower,
		&data.MinActivePower,
		&data.ValidCount,
		&data.AnomalyCount,
	)
	if err != nil {
		// fmt.Printf("Aggregate query error: %v\n", err)
		return nil, nil, err
	}

	// fmt.Printf("Aggregate result - MinKwh: %v, MaxKwh: %v, Peak: %v, Min: %v, Valid: %d, Anomaly: %d\n",
	// 	data.MinTotalImportKwh, data.MaxTotalImportKwh, data.PeakActivePower, data.MinActivePower, data.ValidCount, data.AnomalyCount)

	// Section B: Get Current readings for recalculation
	currentQuery := `
		SELECT 
			metric_value AS current,
			reading_timestamp
		FROM meter_readings_raw
		WHERE client_id = $1
			AND metric_name = 'Current'
			AND validation_status = 'valid'
			AND date_trunc('day', reading_timestamp) >= $2::date
			AND date_trunc('day', reading_timestamp) <= $3::date
		ORDER BY reading_timestamp ASC
	`

	// fmt.Println("=== CURRENT READINGS QUERY ===")
	// fmt.Println(currentQuery)

	rows, err := r.pool.Query(ctx, currentQuery, clientID, startDate, endDate)
	if err != nil {
		// fmt.Printf("Current query error: %v\n", err)
		return nil, nil, err
	}
	defer rows.Close()

	var currentReadings []CurrentReading
	for rows.Next() {
		var reading CurrentReading
		if err := rows.Scan(&reading.Current, &reading.ReadingTime); err != nil {
			return nil, nil, err
		}
		currentReadings = append(currentReadings, reading)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// fmt.Printf("Current readings count: %d\n", len(currentReadings))
	// fmt.Println("==========================================")

	return &data, currentReadings, nil
}

// TimeseriesData represents a single timeseries reading
type TimeseriesData struct {
	MetricName       string
	ReadingTimestamp time.Time
	Value            float64
	Status           string
}

// GetTimeseriesData retrieves meter readings for the last 2 months for specific metrics
func (r *Repository) GetTimeseriesData(ctx context.Context, clientID string) ([]TimeseriesData, error) {
	query := `
		SELECT 
			metric_name,
			reading_timestamp,
			metric_value,
			validation_status
		FROM meter_readings_raw
		WHERE client_id = $1
			AND metric_name IN ('Total_Import_kWh', 'Volts', 'Current', 'Active_Power')
			AND reading_timestamp >= NOW() - INTERVAL '2 months'
		ORDER BY reading_timestamp ASC
	`

	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TimeseriesData
	for rows.Next() {
		var data TimeseriesData
		if err := rows.Scan(&data.MetricName, &data.ReadingTimestamp, &data.Value, &data.Status); err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// UsageData holds aggregated usage data for billing calculation
type UsageData struct {
	MinKwh *float64
	MaxKwh *float64
}

// BillingData holds comprehensive billing data
type BillingData struct {
	MinTotalImportKwh *float64
	MaxTotalImportKwh *float64
	PeakActivePower   *float64
	MinActivePower    *float64
	ValidCount        int
	AnomalyCount      int
}

// CurrentReading represents a current reading
type CurrentReading struct {
	Current     float64
	ReadingTime time.Time
}

// DashboardStats holds dashboard summary statistics
type DashboardStats struct {
	ActiveClientsToday     int
	ActiveClientsYesterday int
	ReadingsToday          int
	ReadingsYesterday      int
	ValidToday             int
	AnomalyToday           int
	InvalidToday           int
}

// GetDashboardStats retrieves dashboard summary statistics
func (r *Repository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	query := `
		WITH time_ranges AS (
			SELECT
				date_trunc('day', now()) AS today_start,
				date_trunc('day', now() - INTERVAL '1 day') AS yesterday_start,
				date_trunc('day', now() + INTERVAL '1 day') AS tomorrow_start
		)
		SELECT
			-- Active clients today (based on last_seen_at)
			(SELECT COUNT(DISTINCT id)
			 FROM meter_clients, time_ranges
			 WHERE date_trunc('day',last_seen_at) = today_start) AS active_clients_today,
			
			-- Active clients yesterday
			(SELECT COUNT(DISTINCT id)
			 FROM meter_clients, time_ranges
			 WHERE date_trunc('day',last_seen_at) = yesterday_start) AS active_clients_yesterday,
			
			-- Readings received today
			(SELECT COUNT(*)
			 FROM meter_readings_raw, time_ranges
			 WHERE date_trunc('day',received_at) = today_start) AS readings_today,
			
			-- Readings received yesterday
			(SELECT COUNT(*)
			 FROM meter_readings_raw, time_ranges
			 WHERE  date_trunc('day',received_at) = yesterday_start) AS readings_yesterday,
			
			-- Valid readings today
			(SELECT COUNT(*)
			 FROM meter_readings_raw, time_ranges
			 WHERE date_trunc('day',received_at) = today_start
			   AND validation_status = 'valid') AS valid_today,
			
			-- Anomaly readings today
			(SELECT COUNT(*)
			 FROM meter_readings_raw, time_ranges
			 WHERE date_trunc('day',received_at) = today_start
			   AND validation_status = 'ANOMALY') AS anomaly_today,
			
			-- Invalid readings today
			(SELECT COUNT(*)
			 FROM meter_readings_raw, time_ranges
			 WHERE date_trunc('day',received_at) = today_start
			   AND validation_status = 'INVALID') AS invalid_today
		FROM time_ranges
		LIMIT 1
	`

	var stats DashboardStats
	err := r.pool.QueryRow(ctx, query).Scan(
		&stats.ActiveClientsToday,
		&stats.ActiveClientsYesterday,
		&stats.ReadingsToday,
		&stats.ReadingsYesterday,
		&stats.ValidToday,
		&stats.AnomalyToday,
		&stats.InvalidToday,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// ClientInfo represents meter client information
type ClientInfo struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
}

// GetAllClients retrieves all meter clients
func (r *Repository) GetAllClients(ctx context.Context) ([]ClientInfo, error) {
	query := `
		SELECT 
			id::text AS client_id,
			client_fingerprint AS name
		FROM meter_clients
		ORDER BY last_seen_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []ClientInfo
	for rows.Next() {
		var client ClientInfo
		if err := rows.Scan(
			&client.ClientID,
			&client.Name,
		); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clients, nil
}

// AnomalyRecord represents an anomaly record from database
type AnomalyRecord struct {
	ID               string
	ClientID         string
	MetricName       string
	MetricValue      float64
	ReadingTimestamp time.Time
	ReceivedAt       time.Time
	ValidationStatus string
	AnomalyReason    *string
}

// GetAnomalies retrieves anomaly records with pagination
func (r *Repository) GetAnomalies(ctx context.Context, startDate, endDate *time.Time, clientID *string, limit, offset int) ([]AnomalyRecord, int, error) {
	// Build WHERE clause
	whereClause := `WHERE validation_status IN ('invalid', 'anomaly')`
	args := []interface{}{}
	paramCount := 0

	// Add date filters if provided
	if startDate != nil {
		paramCount++
		whereClause += fmt.Sprintf(" AND date_trunc('day', reading_timestamp) >= $%d::date", paramCount)
		args = append(args, *startDate)
	}

	if endDate != nil {
		paramCount++
		whereClause += fmt.Sprintf(" AND date_trunc('day', reading_timestamp) < $%d::date", paramCount)
		args = append(args, *endDate)
	}

	// Add client filter if provided
	if clientID != nil && *clientID != "" {
		paramCount++
		whereClause += fmt.Sprintf(" AND client_id = $%d", paramCount)
		args = append(args, *clientID)
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM meter_readings_raw
		%s
	`, whereClause)

	// Log query for debugging
	// fmt.Println("=== COUNT QUERY ===")
	// fmt.Println(countQuery)
	// fmt.Println("=== ARGS ===")
	// for i, arg := range args {
	// 	fmt.Printf("$%d: %v (type: %T)\n", i+1, arg, arg)
	// }
	// fmt.Println("==================")

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// fmt.Printf("=== TOTAL COUNT RESULT: %d ===\n", total)

	// Get paginated data
	paramCount++
	limitParam := fmt.Sprintf("$%d", paramCount)
	paramCount++
	offsetParam := fmt.Sprintf("$%d", paramCount)

	dataQuery := fmt.Sprintf(`
		SELECT 
			id::text,
			client_id::text,
			metric_name,
			metric_value,
			reading_timestamp,
			received_at,
			validation_status,
			anomaly_reason
		FROM meter_readings_raw
		%s
		ORDER BY received_at DESC
		LIMIT %s OFFSET %s
	`, whereClause, limitParam, offsetParam)

	args = append(args, limit, offset)

	// Log data query for debugging
	// fmt.Println("=== DATA QUERY ===")
	// fmt.Println(dataQuery)
	// fmt.Println("=== ARGS ===")
	// for i, arg := range args {
	// 	fmt.Printf("$%d: %v (type: %T)\n", i+1, arg, arg)
	// }
	// fmt.Println("==================")

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []AnomalyRecord
	for rows.Next() {
		var record AnomalyRecord
		if err := rows.Scan(
			&record.ID,
			&record.ClientID,
			&record.MetricName,
			&record.MetricValue,
			&record.ReadingTimestamp,
			&record.ReceivedAt,
			&record.ValidationStatus,
			&record.AnomalyReason,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// fmt.Printf("=== RETURN VALUES ===\n")
	// fmt.Printf("Records count: %d\n", len(records))
	// fmt.Printf("Total: %d\n", total)
	// fmt.Println("=====================")

	return records, total, nil
}
