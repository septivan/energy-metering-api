-- 0002_meter_clients.sql
CREATE TABLE IF NOT EXISTS meter_clients (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_fingerprint TEXT UNIQUE NOT NULL,
  ip_address INET NOT NULL,
  user_agent TEXT,
  first_seen_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_meter_clients_fingerprint ON meter_clients (client_fingerprint);
