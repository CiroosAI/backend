-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    role VARCHAR(20) DEFAULT 'admin',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO admins (username, password, name, email, role, is_active, created_at, updated_at)
VALUES (
    'admin',
    '$2y$10$4gxQkyQlGCvXulvUpRqrV.JW9grktmlcYq8jD8RfUYO884yBsSUY2',
    'Administrator',
    'admin@vladev.xyz',
    'admin',
    true,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    password = VALUES(password),
    name = VALUES(name),
    email = VALUES(email),
    role = VALUES(role),
    is_active = VALUES(is_active),
    updated_at = NOW();