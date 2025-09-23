-- Create refresh_tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id CHAR(64) PRIMARY KEY,
  user_id BIGINT NOT NULL,
  expires_at DATETIME NOT NULL,
  revoked TINYINT(1) DEFAULT 0,
  created_at DATETIME NOT NULL,
  INDEX idx_refresh_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
