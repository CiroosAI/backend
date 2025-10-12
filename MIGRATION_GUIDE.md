# Migration Guide - V2.1 Update

## Masalah: required_vip selalu 0

### Penyebab:
Database yang sedang running belum memiliki kolom `required_vip` dan `purchase_limit` yang baru ditambahkan di schema.

---

## Solusi: Jalankan Migration

### Option 1: Menggunakan MySQL Command Line

```bash
# Login ke MySQL
mysql -u root -p

# Pilih database
USE your_database_name;

# Jalankan migration
SOURCE migrations/add_vip_and_purchase_limit.sql;

# Atau copy-paste isi file migration ke MySQL prompt
```

### Option 2: Menggunakan phpMyAdmin

1. Login ke phpMyAdmin
2. Pilih database Anda
3. Klik tab "SQL"
4. Copy-paste isi file `migrations/add_vip_and_purchase_limit.sql`
5. Klik "Go" / "Execute"

### Option 3: Manual ALTER TABLE

Jalankan query berikut satu per satu:

```sql
-- 1. Tambah kolom required_vip di products
ALTER TABLE `products` 
  ADD COLUMN `required_vip` int DEFAULT '0' 
  COMMENT 'Required VIP level (0 means no requirement)';

-- 2. Tambah kolom purchase_limit di products
ALTER TABLE `products` 
  ADD COLUMN `purchase_limit` int DEFAULT '0' 
  COMMENT 'Maximum purchases per user (0 = unlimited)';

-- 3. Tambah kolom total_invest_vip di users
ALTER TABLE `users` 
  ADD COLUMN `total_invest_vip` decimal(15,2) DEFAULT '0.00' 
  COMMENT 'Total locked category investments for VIP level calculation';

-- 4. Update Insight products (ID 8-12)
UPDATE `products` SET `required_vip` = 1, `purchase_limit` = 1 WHERE `id` = 8;
UPDATE `products` SET `required_vip` = 2, `purchase_limit` = 1 WHERE `id` = 9;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 1 WHERE `id` = 10;
UPDATE `products` SET `required_vip` = 4, `purchase_limit` = 1 WHERE `id` = 11;
UPDATE `products` SET `required_vip` = 5, `purchase_limit` = 1 WHERE `id` = 12;

-- 5. Update AutoPilot products (ID 13-16)
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 2 WHERE `id` = 13;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 2 WHERE `id` = 14;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 1 WHERE `id` = 15;
UPDATE `products` SET `required_vip` = 3, `purchase_limit` = 1 WHERE `id` = 16;

-- 6. Verifikasi
SELECT id, name, required_vip, purchase_limit FROM products ORDER BY id;
```

---

## Verifikasi Setelah Migration

### 1. Cek Kolom Products:
```sql
DESCRIBE products;
```

Harus ada:
- ✅ `required_vip` int
- ✅ `purchase_limit` int

### 2. Cek Kolom Users:
```sql
DESCRIBE users;
```

Harus ada:
- ✅ `total_invest_vip` decimal(15,2)

### 3. Cek Data Products:
```sql
SELECT id, name, required_vip, purchase_limit, category_id 
FROM products 
ORDER BY id;
```

Expected output:
```
ID | Name        | required_vip | purchase_limit | category_id
---|-------------|--------------|----------------|------------
1  | Monitor 1   | 0            | 0              | 1
2  | Monitor 2   | 0            | 0              | 1
...
8  | Insight 1   | 1            | 1              | 2
9  | Insight 2   | 2            | 1              | 2
...
13 | AutoPilot 1 | 3            | 2              | 3
14 | AutoPilot 2 | 3            | 2              | 3
15 | AutoPilot 3 | 3            | 1              | 3
16 | AutoPilot 4 | 3            | 1              | 3
```

---

## Jika Masih Ada Masalah

### Debug Steps:

1. **Cek apakah kolom ada:**
```sql
SHOW COLUMNS FROM products LIKE 'required_vip';
SHOW COLUMNS FROM products LIKE 'purchase_limit';
```

2. **Cek nilai default:**
```sql
SELECT id, name, required_vip, purchase_limit 
FROM products 
WHERE required_vip IS NULL OR purchase_limit IS NULL;
```

Jika ada rows, jalankan:
```sql
UPDATE products SET required_vip = 0 WHERE required_vip IS NULL;
UPDATE products SET purchase_limit = 0 WHERE purchase_limit IS NULL;
```

3. **Restart aplikasi:**
```bash
# Matikan aplikasi yang running
# Lalu start ulang
go run main.go
# atau
docker-compose restart
```

4. **Test API:**
```bash
# GET products
curl http://localhost:8080/api/products

# Cek apakah response memiliki required_vip
```

---

## Alternative: Fresh Database Install

Jika migration terlalu kompleks, bisa install database dari awal:

```bash
# 1. Backup database lama
mysqldump -u root -p your_db_name > backup_old.sql

# 2. Drop database lama (HATI-HATI!)
mysql -u root -p -e "DROP DATABASE your_db_name;"

# 3. Create database baru
mysql -u root -p -e "CREATE DATABASE your_db_name;"

# 4. Import schema baru
mysql -u root -p your_db_name < database/db.sql

# 5. Restore data penting dari backup (users, transactions, dll)
```

---

## Quick Test

Setelah migration, test dengan:

```bash
# Test GET products
curl -X GET http://localhost:8080/api/products

# Cek field required_vip di response
```

Response harus seperti:
```json
{
  "data": {
    "Monitor": [{
      "id": 1,
      "required_vip": 0,      // ← Harus ada!
      "purchase_limit": 0     // ← Harus ada!
    }],
    "Insight": [{
      "id": 8,
      "required_vip": 1,      // ← Harus ada!
      "purchase_limit": 1     // ← Harus ada!
    }]
  }
}
```

---

## Kesimpulan

**Root Cause:** Database belum memiliki kolom baru

**Solution:** Jalankan migration `add_vip_and_purchase_limit.sql`

**Verification:** Query products dan cek field required_vip, purchase_limit muncul

**If Still Fails:** 
1. Check database schema dengan `DESCRIBE products`
2. Restart aplikasi
3. Check logs untuk error
4. Contact development team

---

**Last Updated:** October 12, 2025

