-- +migrate Up
CREATE TABLE IF NOT EXISTS banks (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  code VARCHAR(20) NOT NULL UNIQUE,
  status ENUM('Active','Inactive') NOT NULL DEFAULT 'Active',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Seed common banks/e-wallets
INSERT INTO banks (name, code, status) VALUES
  ('Bank Central Asia', 'BCA', 'Active'),
  ('Bank Rakyat Indonesia', 'BRI', 'Active'),
  ('Bank Negara Indonesia', 'BNI', 'Active'),
  ('Bank Mandiri', 'MANDIRI', 'Active'),
  ('Bank Permata', 'PERMATA', 'Active'),
  ('Bank Neo Commerce', 'BNC', 'Active'),
  ('Dana', 'DANA', 'Active')
ON DUPLICATE KEY UPDATE name = VALUES(name), status = VALUES(status);

-- +migrate Down
DROP TABLE IF EXISTS banks;
