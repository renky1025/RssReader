-- +migrate Up

-- Add email and status fields to users table
ALTER TABLE users ADD COLUMN email TEXT;
ALTER TABLE users ADD COLUMN status INTEGER DEFAULT 1; -- 1=active, 0=disabled
ALTER TABLE users ADD COLUMN last_login_at INTEGER;
ALTER TABLE users ADD COLUMN onboarding_complete INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN last_login_ip TEXT;
ALTER TABLE users ADD COLUMN last_login_device TEXT;

-- Create index for email lookup
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- +migrate Down
-- SQLite doesn't support DROP COLUMN directly, so we'd need to recreate the table
-- For simplicity, we'll leave this as a no-op in down migration
