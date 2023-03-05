BEGIN;

DO $$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'statuses') THEN
      CREATE TYPE statuses AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
   END IF;
END $$;

CREATE TABLE IF NOT EXISTS orders(
   id BIGINT PRIMARY KEY,
   user_id UUID NOT NULL REFERENCES users(id),
   status statuses NOT NULL,
   accrual INT DEFAULT 0,
   created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS user_id_idx ON orders (user_id);

CREATE INDEX IF NOT EXISTS status_idx ON orders (status);

COMMIT;
