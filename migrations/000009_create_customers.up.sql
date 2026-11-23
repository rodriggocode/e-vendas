CREATE TABLE customers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(120) NOT NULL,
  phone VARCHAR(20) NOT NULL,
  email VARCHAR(254),
  tax_id VARCHAR(20),
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ix_customers_phone ON customers (phone);
CREATE UNIQUE INDEX uq_customers_email ON customers (lower(email))
  WHERE email IS NOT NULL AND email <> '';
CREATE UNIQUE INDEX uq_customers_tax_id ON customers (tax_id)
  WHERE tax_id IS NOT NULL AND tax_id <> '';

CREATE TRIGGER trg_customers_updated_at BEFORE UPDATE ON customers
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
