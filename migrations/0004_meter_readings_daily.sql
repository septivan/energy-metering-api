-- 0004_meter_readings_daily.sql
-- Create continuous aggregate for daily readings
CREATE MATERIALIZED VIEW IF NOT EXISTS meter_readings_daily WITH (timescaledb.continuous) AS
SELECT
  time_bucket('1 day', reading_timestamp) AS day,
  client_id,
  metric_name,
  sum(metric_value) AS total_usage,
  avg(metric_value) AS avg_value,
  min(metric_value) AS min_value,
  max(metric_value) AS max_value
FROM meter_readings_raw
GROUP BY day, client_id, metric_name;

-- Refresh policy: refresh continuous aggregate every 1 hour
SELECT add_continuous_aggregate_policy('meter_readings_daily', start_offset => INTERVAL '3 days', end_offset => INTERVAL '1 hour', schedule_interval => INTERVAL '1 hour');
