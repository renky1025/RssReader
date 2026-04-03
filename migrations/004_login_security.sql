-- +migrate Up

-- RSA密钥对存储表
CREATE TABLE IF NOT EXISTS rsa_keys (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key_id TEXT NOT NULL UNIQUE,
  public_key TEXT NOT NULL,
  private_key TEXT NOT NULL,
  created_at INTEGER DEFAULT (strftime('%s','now')),
  expires_at INTEGER NOT NULL,
  is_active INTEGER DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_rsa_keys_key_id ON rsa_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_rsa_keys_active ON rsa_keys(is_active, expires_at);

-- 登录尝试记录表（用于防重放和限流）
CREATE TABLE IF NOT EXISTS login_attempts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  ip_address TEXT NOT NULL,
  success INTEGER DEFAULT 0,
  attempted_at INTEGER DEFAULT (strftime('%s','now')),
  user_agent TEXT,
  nonce TEXT
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts(username, attempted_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip ON login_attempts(ip_address, attempted_at);
CREATE INDEX IF NOT EXISTS idx_login_attempts_nonce ON login_attempts(nonce);

-- +migrate Down
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS rsa_keys;
