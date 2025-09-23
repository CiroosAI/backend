-- Create database (if not exists) and use it
CREATE DATABASE IF NOT EXISTS v1 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE v1;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  number VARCHAR(20) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  reff_code VARCHAR(20) NOT NULL UNIQUE,
  reff_by INT NULL,
  balance DECIMAL(15,2) DEFAULT 0,
  level bigint DEFAULT 0,
  total_invest DECIMAL(15,2) DEFAULT 0,
  spin_ticket bigint DEFAULT 0,
  status ENUM('Active','Inactive','Suspend') DEFAULT 'Active',
  investment_status ENUM('Active','Inactive') DEFAULT 'Inactive',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_users_reff_by (reff_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
