-- +migrate Up
CREATE TABLE IF NOT EXISTS withdrawals (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id INT NOT NULL,
  bank_account_id INT NOT NULL,
  amount DECIMAL(15,2) NOT NULL,
  charge DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  final_amount DECIMAL(15,2) NOT NULL,
  order_id VARCHAR(191) NOT NULL UNIQUE,
  status ENUM('Success','Pending','Failed') NOT NULL DEFAULT 'Pending',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_user_id (user_id),
  INDEX idx_bank_account_id (bank_account_id),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS withdrawals;
