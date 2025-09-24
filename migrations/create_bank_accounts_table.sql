-- +migrate Up
CREATE TABLE IF NOT EXISTS bank_accounts (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id INT NOT NULL,
  bank_id INT UNSIGNED NOT NULL,
  account_name VARCHAR(100) NOT NULL,
  account_number VARCHAR(50) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uniq_user_bank_account (user_id, bank_id, account_number),
  INDEX idx_user_id (user_id),
  INDEX idx_bank_id (bank_id),
  CONSTRAINT fk_bank_accounts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_bank_accounts_bank FOREIGN KEY (bank_id) REFERENCES banks(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS bank_accounts;
