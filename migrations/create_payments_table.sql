-- +migrate Up
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    investment_id INTEGER NOT NULL REFERENCES investments(id) ON DELETE CASCADE,
    reference_id VARCHAR(191),
    order_id VARCHAR(191) NOT NULL UNIQUE,
    payment_method VARCHAR(16),
    payment_channel VARCHAR(16),
    payment_code TEXT,
    payment_link TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'Pending',
    expired_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE IF EXISTS payments;
