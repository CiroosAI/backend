-- Create payment_settings table
CREATE TABLE IF NOT EXISTS payment_settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    pakasir_api_key VARCHAR(191) NOT NULL,
    pakasir_project VARCHAR(191) NOT NULL,
    deposit_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    bank_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(50) NOT NULL,
    account_number VARCHAR(100) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    withdraw_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    wishlist_id TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed one row if empty
INSERT INTO payment_settings (
    pakasir_api_key,
    pakasir_project,
    deposit_amount,
    bank_name,
    bank_code,
    account_number,
    account_name,
    withdraw_amount,
    wishlist_id
) VALUES (
    'AWD1A2AWD132',
    'AWD1SAD2A1W',
    10000.00,
    'Bank BCA',
    'BCA',
    '1234567890',
    'StoneForm Admin',
    50000.00,
    '1'
) ON DUPLICATE KEY UPDATE
    pakasir_api_key = VALUES(pakasir_api_key),
    pakasir_project = VALUES(pakasir_project),
    deposit_amount = VALUES(deposit_amount),
    bank_name = VALUES(bank_name),
    bank_code = VALUES(bank_code),
    account_number = VALUES(account_number),
    account_name = VALUES(account_name),
    withdraw_amount = VALUES(withdraw_amount),
    wishlist_id = VALUES(wishlist_id);