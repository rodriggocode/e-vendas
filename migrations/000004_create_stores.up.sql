CREATE TABLE stores (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(120) NOT NULL,
  description TEXT,
  whatsapp VARCHAR(20),
  address_line VARCHAR(255),
  neighborhood VARCHAR(120),
  city VARCHAR(120),
  state CHAR(2),
  logo_url TEXT,
  opening_hours JSONB,
  accepts_pickup BOOLEAN NOT NULL DEFAULT FALSE,
  accepts_delivery BOOLEAN NOT NULL DEFAULT FALSE,
  delivery_fee_cents BIGINT NOT NULL DEFAULT 0,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  rejection_reason TEXT,
  min_order_cents BIGINT NOT NULL DEFAULT 0,
  prep_time_min INTEGER,
  delivery_time_min INTEGER,
  pix_key VARCHAR(255),
  payment_instructions TEXT,
  accepted_payment_methods TEXT[] NOT NULL DEFAULT ARRAY['pix'],
  paused BOOLEAN NOT NULL DEFAULT FALSE,
  order_seq BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE stores ADD CONSTRAINT ck_stores_slug CHECK (slug ~ '^[a-z0-9-]+$');
ALTER TABLE stores ADD CONSTRAINT ck_stores_status CHECK (status IN ('pending', 'active', 'suspended', 'rejected'));
ALTER TABLE stores ADD CONSTRAINT ck_stores_rejection_reason CHECK (status = 'rejected' OR rejection_reason IS NULL);
ALTER TABLE stores ADD CONSTRAINT ck_stores_payment_methods
  CHECK (cardinality(accepted_payment_methods) > 0
         AND accepted_payment_methods <@ ARRAY['pix', 'cash', 'card_on_delivery']::text[]);
ALTER TABLE stores ADD CONSTRAINT ck_stores_fulfillment CHECK (accepts_pickup OR accepts_delivery);
ALTER TABLE stores ADD CONSTRAINT ck_stores_delivery_fee CHECK (delivery_fee_cents >= 0);
ALTER TABLE stores ADD CONSTRAINT ck_stores_min_order CHECK (min_order_cents >= 0);
ALTER TABLE stores ADD CONSTRAINT ck_stores_order_seq CHECK (order_seq >= 0);

CREATE UNIQUE INDEX uq_stores_slug ON stores (slug);
CREATE INDEX ix_stores_city_status ON stores (city, status);

CREATE TRIGGER trg_stores_updated_at BEFORE UPDATE ON stores
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
