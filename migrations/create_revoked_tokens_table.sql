-- Create revocation table for access-token jtis (DB fallback)
CREATE TABLE IF NOT EXISTS revoked_tokens (
  id VARCHAR(128) NOT NULL PRIMARY KEY,
  revoked_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
