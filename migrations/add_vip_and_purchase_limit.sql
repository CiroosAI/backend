-- Migration: Add VIP and Purchase Limit Features
-- Run this on your existing database

-- Add required_vip column to products table
ALTER TABLE `products` 
  ADD COLUMN IF NOT EXISTS `required_vip` int DEFAULT '0' 
  COMMENT 'Required VIP level (0 means no requirement)';

-- Add purchase_limit column to products table
ALTER TABLE `products` 
  ADD COLUMN IF NOT EXISTS `purchase_limit` int DEFAULT '0' 
  COMMENT 'Maximum purchases per user (0 = unlimited)';

-- Add total_invest_vip column to users table
ALTER TABLE `users` 
  ADD COLUMN IF NOT EXISTS `total_invest_vip` decimal(15,2) DEFAULT '0.00' 
  COMMENT 'Total locked category investments for VIP level calculation';

-- Update existing products with default values if needed
UPDATE `products` SET `required_vip` = 0 WHERE `required_vip` IS NULL;
UPDATE `products` SET `purchase_limit` = 0 WHERE `purchase_limit` IS NULL;
UPDATE `users` SET `total_invest_vip` = 0.00 WHERE `total_invest_vip` IS NULL;

-- Update Insight products with VIP requirements and purchase limit
UPDATE `products` SET `required_vip` = 1, `purchase_limit` = 1 WHERE `id` = 8;
UPDATE `products` SET `required_vip` = 2, `purchase_limit` = 1 WHERE `id` = 9;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 1 WHERE `id` = 10;
UPDATE `products` SET `required_vip` = 4, `purchase_limit` = 1 WHERE `id` = 11;
UPDATE `products` SET `required_vip` = 5, `purchase_limit` = 1 WHERE `id` = 12;

-- Update AutoPilot products with VIP requirements and purchase limit
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 2 WHERE `id` = 13;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 2 WHERE `id` = 14;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 1 WHERE `id` = 15;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 1 WHERE `id` = 16;

-- Monitor products remain unlimited (purchase_limit = 0)
-- They already have default values

-- Add index for better performance
CREATE INDEX IF NOT EXISTS `idx_products_required_vip` ON `products` (`required_vip`);
CREATE INDEX IF NOT EXISTS `idx_products_purchase_limit` ON `products` (`purchase_limit`);

-- Verify changes
SELECT id, name, required_vip, purchase_limit FROM products ORDER BY id;

