-- 1. Spin Prizes Table (Daftar hadiah yang tersedia)
CREATE TABLE IF NOT EXISTS spin_prizes (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  amount DECIMAL(15,2) NOT NULL,
  code VARCHAR(20) NOT NULL UNIQUE COMMENT 'Unique code untuk validasi claim prize',
  chance_weight INT NOT NULL COMMENT 'Weight untuk random selection (semakin besar semakin sering muncul)',
  status ENUM('Active','Inactive') NOT NULL DEFAULT 'Active',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  INDEX idx_status (status),
  INDEX idx_code (code),
  INDEX idx_chance_weight (chance_weight)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Available spin wheel prizes';

-- 2. User Spins Table (History spin user)
CREATE TABLE IF NOT EXISTS user_spins (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id INT NOT NULL,
  spin_prize_id INT UNSIGNED NOT NULL COMMENT 'Reference to won prize',
  amount DECIMAL(15,2) NOT NULL COMMENT 'Amount yang dimenangkan',
  prize_code VARCHAR(20) NOT NULL COMMENT 'Code hadiah yang dimenangkan',
  next_spin_at DATETIME NOT NULL COMMENT 'Kapan user bisa spin lagi (created_at + 24 jam)',
  status ENUM('Pending','Claimed','Failed') NOT NULL DEFAULT 'Pending' COMMENT 'Status claim hadiah',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  INDEX idx_user_id (user_id),
  INDEX idx_next_spin_at (next_spin_at),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at),
  INDEX idx_user_next_spin (user_id, next_spin_at), -- Composite index untuk query cek last spin

  CONSTRAINT fk_user_spins_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_spins_prize FOREIGN KEY (spin_prize_id) REFERENCES spin_prizes(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User spin wheel history and claims';

-- Insert sample spin prizes dengan weighted chance system
INSERT INTO spin_prizes (amount, code, chance_weight, status) VALUES
(1000.00, 'SPIN_1K', 5000, 'Active'),
(5000.00, 'SPIN_5K', 2000, 'Active'),
(10000.00, 'SPIN_10K', 500, 'Active'),
(50000.00, 'SPIN_50K', 80, 'Active'),
(100000.00, 'SPIN_100K', 10, 'Active'),
(200000.00, 'SPIN_200K', 3, 'Active'),
(500000.00, 'SPIN_500K', 2, 'Active'),
(1000000.00, 'SPIN_1M', 1, 'Active')
ON DUPLICATE KEY UPDATE code = VALUES(code);

-- Trigger untuk auto-set next_spin_at
DELIMITER $$
CREATE TRIGGER user_spins_set_next_spin
BEFORE INSERT ON user_spins
FOR EACH ROW
BEGIN
    IF NEW.next_spin_at IS NULL THEN
        SET NEW.next_spin_at = DATE_ADD(NOW(), INTERVAL 24 HOUR);
    END IF;
END$$
DELIMITER ;

-- View untuk memudahkan query active prizes dengan percentage
DROP VIEW IF EXISTS spin_prizes_with_percentage;
CREATE VIEW spin_prizes_with_percentage AS
SELECT 
    id,
    amount,
    code,
    chance_weight,
    ROUND((chance_weight * 100.0) / (SELECT SUM(chance_weight) FROM spin_prizes WHERE status = 'Active'), 2) as chance_percentage,
    status
FROM spin_prizes 
WHERE status = 'Active'
ORDER BY amount ASC;
