INSERT INTO products (name, minimum, maximum, percentage, duration, status)
VALUES
  ('Bintang 1', 30000.00, 1000000.00, 15.00, 200, 'Active'),
  ('Bintang 2', 1500000.00, 3000000.00, 30.00, 67, 'Active'),
  ('Bintang 3', 5000000.00, 10000000.00, 40.00, 40, 'Active')
ON DUPLICATE KEY UPDATE
  minimum = VALUES(minimum),
  maximum = VALUES(maximum),
  percentage = VALUES(percentage),
  duration = VALUES(duration),
  status = VALUES(status);