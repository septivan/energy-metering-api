-- 0001_enable_extensions.sql
-- Enable TimescaleDB and pgcrypto
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS timescaledb;
