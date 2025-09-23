CREATE TABLE IF NOT EXISTS transactions (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    charge DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    order_id VARCHAR(191) NOT NULL,
    transaction_flow ENUM('debit','credit') NOT NULL COMMENT 'debit=money in, credit=money out',
    transaction_type VARCHAR(50) NOT NULL COMMENT 'deposit, withdraw, return, bonus, etc',
    message TEXT NULL,
    status ENUM('Success','Pending','Failed') NOT NULL DEFAULT 'Pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    INDEX idx_user_id (user_id),
    INDEX idx_order_id (order_id),
    INDEX idx_transaction_flow (transaction_flow),
    INDEX idx_transaction_type (transaction_type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),

    CONSTRAINT fk_transactions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User transaction records';

-- Add unique constraint for order_id to prevent duplicate transactions
ALTER TABLE transactions ADD UNIQUE KEY unique_order_id (order_id);

-- Optional: Create index for common queries
CREATE INDEX idx_user_status_created ON transactions (user_id, status, created_at);
CREATE INDEX idx_user_type_created ON transactions (user_id, transaction_type, created_at);