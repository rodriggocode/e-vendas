CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  store_id UUID NOT NULL REFERENCES stores (id) ON DELETE RESTRICT,
  category_id UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
  name VARCHAR(120) NOT NULL,
  description TEXT,
  sku VARCHAR(64),
  price_cents BIGINT NOT NULL,
  stock_quantity INTEGER,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE products ADD CONSTRAINT ck_products_price CHECK (price_cents >= 0);
ALTER TABLE products ADD CONSTRAINT ck_products_stock CHECK (stock_quantity IS NULL OR stock_quantity >= 0);

CREATE UNIQUE INDEX uq_products_store_sku ON products (store_id, sku);
CREATE INDEX ix_products_store_active ON products (store_id, active);
CREATE INDEX ix_products_category ON products (category_id);
CREATE UNIQUE INDEX uq_products_id_store ON products (id, store_id);

CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
