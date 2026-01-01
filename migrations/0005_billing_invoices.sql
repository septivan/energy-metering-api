-- 0005_billing_invoices.sql
CREATE TABLE IF NOT EXISTS billing_invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id UUID REFERENCES meter_clients(id),
  billing_period_start DATE,
  billing_period_end DATE,
  total_usage DOUBLE PRECISION,
  total_amount NUMERIC(12,2),
  currency TEXT DEFAULT 'USD',
  pdf_url TEXT,
  generated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_billing_generated_at ON billing_invoices (generated_at DESC);
