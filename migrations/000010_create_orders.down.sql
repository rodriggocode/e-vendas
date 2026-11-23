DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS trg_orders_assign_number ON orders;
DROP FUNCTION IF EXISTS orders_assign_number();

DO $$
DECLARE s text;
BEGIN
  FOR s IN SELECT sequencename FROM pg_sequences
           WHERE schemaname = current_schema() AND sequencename LIKE 'orders\_number\_seq\_%'
  LOOP
    EXECUTE format('DROP SEQUENCE IF EXISTS %I', s);
  END LOOP;
END $$;

DROP TABLE orders;
