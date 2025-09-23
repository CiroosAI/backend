-- Create payment_settings table
CREATE TABLE IF NOT EXISTS payment_settings (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  pakasir_api_key VARCHAR(191) NOT NULL,
  pakasir_project VARCHAR(191) NOT NULL,
  deposit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  bank_name VARCHAR(100) NOT NULL,
  bank_code VARCHAR(50) NOT NULL,
  account_number VARCHAR(100) NOT NULL,
  account_name VARCHAR(100) NOT NULL,
  withdraw_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
  wishlist_id TEXT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed one row if empty
INSERT INTO payment_settings (pakasir_api_key, pakasir_project, deposit_amount, bank_name, bank_code, account_number, account_name, withdraw_amount, wishlist_id)
SELECT '', '', 0.00, '', '', '', '', 0.00, NULL
WHERE NOT EXISTS (SELECT 1 FROM payment_settings);
