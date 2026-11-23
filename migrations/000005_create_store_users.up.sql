CREATE TABLE store_users (
  store_id UUID NOT NULL REFERENCES stores (id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  role VARCHAR(20) NOT NULL DEFAULT 'staff',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (store_id, user_id)
);

ALTER TABLE store_users ADD CONSTRAINT ck_store_users_role CHECK (role IN ('owner', 'staff'));
CREATE INDEX ix_store_users_user_id ON store_users (user_id);
