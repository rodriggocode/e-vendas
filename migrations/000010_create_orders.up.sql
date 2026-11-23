CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  store_id UUID NOT NULL REFERENCES stores (id) ON DELETE RESTRICT,
  customer_id UUID NOT NULL REFERENCES customers (id) ON DELETE RESTRICT,
  number BIGINT,
  tracking_code VARCHAR(64) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  fulfillment_type VARCHAR(20) NOT NULL,
  payment_method VARCHAR(20) NOT NULL,
  payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
  paid_at TIMESTAMPTZ,
  subtotal_cents BIGINT NOT NULL,
  discount_cents BIGINT NOT NULL DEFAULT 0,
  shipping_cents BIGINT NOT NULL DEFAULT 0,
  total_cents BIGINT NOT NULL,
  notes TEXT,
  shipping_address JSONB,
  change_for_cents BIGINT,
  cancel_reason TEXT,
  confirmed_at TIMESTAMPTZ,
  preparing_at TIMESTAMPTZ,
  ready_at TIMESTAMPTZ,
  out_for_delivery_at TIMESTAMPTZ,
  delivered_at TIMESTAMPTZ,
  cancelled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE orders ADD CONSTRAINT ck_orders_status
  CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'out_for_delivery', 'delivered', 'cancelled'));
ALTER TABLE orders ADD CONSTRAINT ck_orders_fulfillment CHECK (fulfillment_type IN ('pickup', 'delivery'));
ALTER TABLE orders ADD CONSTRAINT ck_orders_payment_method CHECK (payment_method IN ('pix', 'cash', 'card_on_delivery'));
ALTER TABLE orders ADD CONSTRAINT ck_orders_payment_status CHECK (payment_status IN ('pending', 'paid', 'cancelled'));
ALTER TABLE orders ADD CONSTRAINT ck_orders_number CHECK (number IS NULL OR number > 0);
ALTER TABLE orders ADD CONSTRAINT ck_orders_amounts
  CHECK (subtotal_cents >= 0 AND discount_cents >= 0 AND shipping_cents >= 0 AND total_cents >= 0);
ALTER TABLE orders ADD CONSTRAINT ck_orders_total
  CHECK (total_cents = subtotal_cents - discount_cents + shipping_cents);
ALTER TABLE orders ADD CONSTRAINT ck_orders_change
  CHECK (change_for_cents IS NULL OR (payment_method = 'cash' AND change_for_cents >= total_cents));
ALTER TABLE orders ADD CONSTRAINT ck_orders_cancel_reason CHECK (status = 'cancelled' OR cancel_reason IS NULL);
ALTER TABLE orders ADD CONSTRAINT ck_orders_delivery_address
  CHECK (fulfillment_type <> 'delivery' OR shipping_address IS NOT NULL);
ALTER TABLE orders ADD CONSTRAINT ck_orders_tracking_code
  CHECK (char_length(tracking_code) >= 32);

CREATE UNIQUE INDEX uq_orders_store_number ON orders (store_id, number);
CREATE UNIQUE INDEX uq_orders_tracking_code ON orders (tracking_code);
CREATE INDEX ix_orders_store_status ON orders (store_id, status);
CREATE INDEX ix_orders_store_created ON orders (store_id, created_at DESC);
CREATE INDEX ix_orders_customer ON orders (customer_id);
CREATE UNIQUE INDEX uq_orders_id_store ON orders (id, store_id);

-- numero do pedido: sequencial por loja, gerado no banco (BANCO-DE-DADOS.md, premissa 11).
-- Usa uma sequence por loja em vez de UPDATE em stores, para nao inflar stores.updated_at.
CREATE OR REPLACE FUNCTION orders_assign_number()
RETURNS TRIGGER AS $$
DECLARE seq_name text;
BEGIN
  IF NEW.number IS NULL THEN
    seq_name := 'orders_number_seq_' || replace(NEW.store_id::text, '-', '_');
    -- duas primeiras compras simultâneas da mesma loja disputam a criação da
    -- sequence; o advisory lock serializa e evita 23505 em pg_class (M-G)
    PERFORM pg_advisory_xact_lock(hashtext(seq_name));
    IF to_regclass(seq_name) IS NULL THEN
      EXECUTE format('CREATE SEQUENCE %I', seq_name);
    END IF;
    NEW.number := nextval(seq_name);
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_orders_assign_number BEFORE INSERT ON orders
  FOR EACH ROW EXECUTE FUNCTION orders_assign_number();

CREATE TRIGGER trg_orders_updated_at BEFORE UPDATE ON orders
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
