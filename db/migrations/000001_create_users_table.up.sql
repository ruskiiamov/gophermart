BEGIN;

CREATE TABLE IF NOT EXISTS users(
   id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
   login VARCHAR (50) UNIQUE NOT NULL,
   pass_hash VARCHAR (50) NOT NULL,
   access_token VARCHAR (50) UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS login_idx ON users (login);

CREATE INDEX IF NOT EXISTS access_token_idx ON users (access_token);

COMMIT;
