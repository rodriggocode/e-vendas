DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS trg_orders_assign_number ON orders;
DROP FUNCTION IF EXISTS orders_assign_number();
DROP TABLE orders;
