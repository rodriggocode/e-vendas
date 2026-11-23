CREATE TABLE order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL,
  product_id UUID NOT NULL,
  store_id UUID NOT NULL,
  product_name VARCHAR(120) NOT NULL,
  quantity INTEGER NOT NULL,
  notes VARCHAR(500),
  unit_price_cents BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (order_id, store_id) REFERENCES orders (id, store_id) ON DELETE CASCADE,
  FOREIGN KEY (product_id, store_id) REFERENCES products (id, store_id) ON DELETE RESTRICT
);

ALTER TABLE order_items ADD CONSTRAINT ck_order_items_quantity CHECK (quantity > 0);
ALTER TABLE order_items ADD CONSTRAINT ck_order_items_price CHECK (unit_price_cents >= 0);
CREATE UNIQUE INDEX uq_order_items_order_product ON order_items (order_id, product_id);
