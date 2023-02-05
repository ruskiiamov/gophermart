BEGIN;

CREATE TABLE IF NOT EXISTS withdrawals(
   id BIGINT PRIMARY KEY,
   user_id UUID NOT NULL REFERENCES users(id),
   sum INT NOT NULL,
   created_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS user_id_idx ON withdrawals (user_id);

COMMIT;
