# CIROOS V2 - Investment System Overhaul Documentation

## Overview
The investment system has been completely revamped from an open-amount system to a fixed-product system with **dynamic categories** (not hardcoded) and a VIP level requirement mechanism. Admin can now manage categories and products through API.

---

## Major Changes Summary

### 1. **Dynamic Category System** ⭐ NEW
- Categories are now **database-driven**, not hardcoded enum
- Admin can create, edit, and delete categories via API
- Each category has configurable `profit_type` (locked/unlocked)
- Default categories: Monitor (locked), Insight (unlocked), AutoPilot (unlocked)

### 2. **Product System Transformation**
- **OLD**: 3 open-amount products (Bintang 1, 2, 3) where users could input any amount within min/max ranges
- **NEW**: 16 fixed-amount products across dynamic categories
- Products are linked to categories via `category_id` (foreign key)
- Admin can add products via POST API

### 3. **VIP Level System**
- VIP levels calculated based on **total investments in "locked" profit type categories**
- VIP requirements configurable per product
- Automatic VIP level upgrades when investment thresholds are reached

### 4. **Simplified Bonus System** ⭐ CHANGED
- **OLD**: Multi-level team bonuses (5%, 2%, 1% for levels 1-3)
- **NEW**: Single-level referral bonus only - **30% for direct referrer (level 1)**
- Team management bonuses completely removed
- No more complex team hierarchy bonus calculations

### 5. **Profit Distribution Model**
- **Locked Profit** (e.g., Monitor): Profits accumulate but only paid to balance when investment completes
- **Unlocked Profit** (e.g., Insight/AutoPilot): Profits paid immediately when investment completes

---

## Database Schema Changes

### NEW: Categories Table ⭐
**Completely New Table** for dynamic category management:

```sql
CREATE TABLE `categories` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(100) NOT NULL,
  `description` text,
  `profit_type` enum('locked','unlocked') NOT NULL DEFAULT 'unlocked',
  `status` enum('Active','Inactive') NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Default Data:**
- Monitor (id=1, profit_type='locked')
- Insight (id=2, profit_type='unlocked')
- AutoPilot (id=3, profit_type='unlocked')

### Products Table
**Changed Fields:**
- ❌ Removed: `minimum`, `maximum`, `percentage`, `category` (enum)
- ✅ Added: `category_id` (foreign key), `amount`, `daily_profit`, `required_vip`

```sql
CREATE TABLE `products` (
  `id` int UNSIGNED NOT NULL,
  `category_id` int UNSIGNED NOT NULL,  -- Foreign key to categories
  `name` varchar(100) NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `daily_profit` decimal(15,2) NOT NULL,
  `duration` int NOT NULL,
  `required_vip` int DEFAULT '0',
  `status` enum('Active','Inactive') NOT NULL DEFAULT 'Active',
  FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT
);
```

### Investments Table
**Changed Fields:**
- ❌ Removed: `percentage`, `category` (enum)
- ✅ Added: `category_id` (foreign key)

```sql
CREATE TABLE `investments` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `product_id` int UNSIGNED NOT NULL,
  `category_id` int UNSIGNED NOT NULL,  -- Foreign key to categories
  `amount` decimal(15,2) NOT NULL,
  `daily_profit` decimal(15,2) NOT NULL,
  `duration` int NOT NULL,
  `total_paid` int NOT NULL DEFAULT '0',
  `total_returned` decimal(15,2) NOT NULL DEFAULT '0.00',
  FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT
);
```

### Users Table
**Added Field:**
- ✅ `total_invest_vip` - Tracks "locked" category investments for VIP calculation

```sql
ALTER TABLE `users` ADD COLUMN 
  `total_invest_vip` decimal(15,2) DEFAULT '0.00' 
  COMMENT 'Total locked category investments for VIP level calculation';
```

**Important - Two Investment Tracking Fields:**
1. **`total_invest`**: Tracks ALL investments (Monitor + Insight + AutoPilot)
   - Used to display total investment history to user
   - Shows how much user has invested overall
   
2. **`total_invest_vip`**: Tracks ONLY "locked" profit type category investments
   - Used ONLY for VIP level calculation
   - Only Monitor (locked) category counts
   - Insight and AutoPilot do NOT increase this value
   
3. **`level`**: VIP level (0-5)
   - Calculated automatically from `total_invest_vip`
   - Used to validate product purchase eligibility

---

## Product Categories & Details

### 1. Monitor Category (Locked Profit) - 7 Products

| Product ID | Name | Amount | Daily Profit | Duration | VIP Required | Purchase Limit |
|------------|------|--------|--------------|----------|--------------|----------------|
| 1 | Monitor 1 | 50,000 | 15,000 | 70 days | 0 | Unlimited |
| 2 | Monitor 2 | 200,000 | 68,000 | 60 days | 0 | Unlimited |
| 3 | Monitor 3 | 500,000 | 175,000 | 65 days | 0 | Unlimited |
| 4 | Monitor 4 | 1,250,000 | 432,000 | 65 days | 0 | Unlimited |
| 5 | Monitor 5 | 2,800,000 | 1,050,000 | 65 days | 0 | Unlimited |
| 6 | Monitor 6 | 7,000,000 | 2,660,000 | 50 days | 0 | Unlimited |
| 7 | Monitor 7 | 20,000,000 | 8,000,000 | 50 days | 0 | Unlimited |

**Key Behavior:**
- Daily profits accumulate in `total_returned` field
- Balance is NOT updated daily
- When investment completes, total profit = `daily_profit × duration` is paid to user balance
- Example: Monitor 1 pays 15,000 × 70 = 1,050,000 at completion
- **No purchase limit** - users can buy Monitor products unlimited times
- **Only Monitor investments count toward VIP level**

### 2. Insight Category (Unlocked) - 5 Products ⭐ LIMITED

| Product ID | Name | Amount | Daily Profit | Total Return | Duration | VIP Required | Purchase Limit |
|------------|------|--------|--------------|--------------|----------|--------------|----------------|
| 8 | Insight 1 | 50,000 | 20,000 | 70,000 | 1 day | VIP 1 | **1x only** |
| 9 | Insight 2 | 250,000 | 275,000 | 525,000 | 1 day | VIP 2 | **1x only** |
| 10 | Insight 3 | 700,000 | 950,000 | 1,650,000 | 1 day | VIP 3 | **1x only** |
| 11 | Insight 4 | 2,000,000 | 3,600,000 | 5,600,000 | 1 day | VIP 4 | **1x only** |
| 12 | Insight 5 | 8,000,000 | 16,000,000 | 24,000,000 | 1 day | VIP 5 | **1x only** |

**Key Behavior:**
- Profits paid to balance when investment completes (1 day)
- Requires specific VIP level to purchase
- **Each product limited to 1 purchase per user LIFETIME**
- Once purchased, cannot buy the same Insight product again
- Total return = Amount + Daily Profit

### 3. AutoPilot Category - 4 Products ⭐ LIMITED

| Product ID | Name | Amount | Daily Profit | Total Return | Duration | VIP Required | Purchase Limit |
|------------|------|--------|--------------|--------------|----------|--------------|----------------|
| 13 | AutoPilot 1 | 80,000 | 70,000 | 150,000 | 1 day | VIP 3 | **2x only** |
| 14 | AutoPilot 2 | 165,000 | 150,000 | 315,000 | 1 day | VIP 3 | **2x only** |
| 15 | AutoPilot 3 | 750,000 | 1,000,000 | 1,750,000 | 1 day | VIP 3 | **1x only** |
| 16 | AutoPilot 4 | 2,450,000 | 4,000,000 | 6,450,000 | 1 day | VIP 3 | **1x only** |

**Key Behavior:**
- All require VIP 3
- Profits paid to balance when investment completes (1 day)
- **AutoPilot 1 & 2**: Can purchase 2 times lifetime
- **AutoPilot 3 & 4**: Can purchase 1 time lifetime only
- Total return = Amount + Daily Profit

---

## VIP Level System

### VIP Thresholds (Based on Locked Category Investments)

| VIP Level | Minimum Investment | Unlocks |
|-----------|-------------------|---------|
| VIP 0 | 0 | Monitor products only |
| VIP 1 | 50,000 | Insight 1 |
| VIP 2 | 1,200,000 | Insight 2 |
| VIP 3 | 7,000,000 | Insight 3, All AutoPilot |
| VIP 4 | 30,000,000 | Insight 4 |
| VIP 5 | 150,000,000 | Insight 5 |

### VIP Calculation Logic
```go
func calculateVIPLevel(totalInvestVIP float64) uint {
    if totalInvestVIP >= 150000000 { return 5 }
    else if totalInvestVIP >= 30000000 { return 4 }
    else if totalInvestVIP >= 7000000 { return 3 }
    else if totalInvestVIP >= 1200000 { return 2 }
    else if totalInvestVIP >= 50000 { return 1 }
    return 0
}
```

### VIP Validation
**Product Purchase Validation:**
```go
// User level must be >= product's required_vip
if userLevel < product.RequiredVIP {
    return error("VIP level tidak cukup")
}
```

**Notes:**
- VIP level stored in `users.level` field (0-5)
- VIP level is automatically updated when locked category investment is confirmed
- Only "locked" profit_type categories count toward VIP level (tracked in `total_invest_vip`)
- Insight and AutoPilot do NOT increase VIP level
- Product purchase checks `users.level >= products.required_vip`

### Frontend VIP Display
- Dashboard shows VIP card with progress bar
- `/vip` page shows detailed timeline and benefits
- Products show lock icon if VIP requirement not met
- Buy button disabled if `user.level < product.required_vip`

---

## New Admin APIs ⭐

### Admin Categories Management

#### 1. GET /api/admin/categories
List all categories (Active and Inactive)

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "categories": [
      {
        "id": 1,
        "name": "Monitor",
        "description": "Profit terkunci, dibayarkan saat investasi selesai",
        "profit_type": "locked",
        "status": "Active",
        "created_at": "2025-10-11T00:00:00Z",
        "updated_at": "2025-10-11T00:00:00Z"
      }
    ]
  }
}
```

#### 2. GET /api/admin/categories/{id}
Get single category details

#### 3. POST /api/admin/categories
Create new category

**Request:**
```json
{
  "name": "Premium",
  "description": "Premium investment category",
  "profit_type": "locked",
  "status": "Active"
}
```

**Validation:**
- `name`: Required
- `profit_type`: Must be "locked" or "unlocked" (default: "unlocked")
- `status`: Must be "Active" or "Inactive" (default: "Active")

#### 4. PUT /api/admin/categories/{id}
Update category (can change name via admin!)

**Request:**
```json
{
  "name": "Monitor Plus",
  "description": "Updated description",
  "profit_type": "locked",
  "status": "Active"
}
```

#### 5. DELETE /api/admin/categories/{id}
Delete category (prevented if products exist)

**Error Response:**
```json
{
  "success": false,
  "message": "Tidak dapat menghapus kategori yang masih digunakan oleh produk"
}
```

---

### Admin Products Management

#### 1. GET /api/admin/products
List all products with category information

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "products": [
      {
        "id": 1,
        "category_id": 1,
        "name": "Monitor 1",
        "amount": 50000,
        "daily_profit": 15000,
        "duration": 70,
        "required_vip": 0,
        "status": "Active",
        "created_at": "2025-10-11T00:00:00Z",
        "updated_at": "2025-10-11T00:00:00Z",
        "category": {
          "id": 1,
          "name": "Monitor",
          "profit_type": "locked",
          "status": "Active"
        }
      }
    ]
  }
}
```

#### 2. GET /api/admin/products/{id}
Get single product with category

#### 3. POST /api/admin/products ⭐ NEW
Create new product

**Request:**
```json
{
  "category_id": 1,
  "name": "Monitor 8",
  "amount": 50000000,
  "daily_profit": 20000000,
  "duration": 45,
  "required_vip": 0,
  "purchase_limit": 0,
  "status": "Active"
}
```

**Validation:**
- `category_id`: Required, must exist in categories table
- `name`: Required
- `amount`: Required, must be > 0
- `daily_profit`: Required, must be > 0
- `duration`: Required, must be > 0
- `required_vip`: Optional (default: 0)
- `purchase_limit`: Optional (default: 0, means unlimited)
- `status`: Optional (default: "Active")

**Purchase Limit:**
- `0` = Unlimited purchases (typically for Monitor)
- `1` = User can only buy once (Insight products)
- `2` = User can buy 2 times (AutoPilot 1 & 2)

**Response:**
```json
{
  "success": true,
  "message": "Produk berhasil dibuat",
  "data": {
    "id": 17,
    "category_id": 1,
    "name": "Monitor 8",
    "amount": 50000000,
    "daily_profit": 20000000,
    "duration": 45,
    "required_vip": 0,
    "status": "Active",
    "category": {
      "id": 1,
      "name": "Monitor"
    }
  }
}
```

#### 4. PUT /api/admin/products/{id}
Update product

**Request (all fields optional):**
```json
{
  "category_id": 2,
  "name": "Insight 6",
  "amount": 15000000,
  "daily_profit": 45000000,
  "duration": 1,
  "required_vip": 5,
  "status": "Inactive"
}
```

#### 5. DELETE /api/admin/products/{id}
Delete product (prevented if investments exist)

---

## User API Changes

### 1. GET /api/products
**Description:** Get all active products grouped by category

**Request:**
```http
GET /api/products HTTP/1.1
Host: api.example.com
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "Monitor": [
      {
        "id": 1,
        "name": "Monitor 1",
        "category": "Monitor",
        "amount": 50000,
        "daily_profit": 15000,
        "duration": 70,
        "required_vip": 0,
        "status": "Active",
        "created_at": "2025-10-11T00:00:00Z",
        "updated_at": "2025-10-11T00:00:00Z"
      },
      // ... more Monitor products
    ],
    "Insight": [
      {
        "id": 8,
        "name": "Insight 1",
        "category": "Insight",
        "amount": 50000,
        "daily_profit": 70000,
        "duration": 1,
        "required_vip": 1,
        "status": "Active",
        "created_at": "2025-10-11T00:00:00Z",
        "updated_at": "2025-10-11T00:00:00Z"
      },
      // ... more Insight products
    ],
    "AutoPilot": [
      {
        "id": 13,
        "name": "AutoPilot 1",
        "category": "AutoPilot",
        "amount": 80000,
        "daily_profit": 70000,
        "duration": 1,
        "required_vip": 3,
        "status": "Active",
        "created_at": "2025-10-11T00:00:00Z",
        "updated_at": "2025-10-11T00:00:00Z"
      },
      // ... more AutoPilot products
    ]
  }
}
```

---

### 2. POST /api/users/investments
**Description:** Create a new investment (amount is now automatic from product)

**Changes:**
- ❌ Removed: `amount` field from request
- ✅ Product's fixed amount is used automatically
- ✅ VIP level validation for Insight/AutoPilot

**Request:**
```json
{
  "product_id": 1,
  "payment_method": "QRIS",
  "payment_channel": ""
}
```

**OLD Request (no longer valid):**
```json
{
  "product_id": 1,
  "amount": 100000,  // ❌ This field is removed
  "payment_method": "QRIS",
  "payment_channel": ""
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Pembelian berhasil, silakan lakukan pembayaran",
  "data": {
    "order_id": "INV1234567890",
    "amount": 50000,
    "product": "Monitor 1",
    "category": "Monitor",
    "duration": 70,
    "daily_profit": 15000,
    "status": "Pending"
  }
}
```

**Response (VIP Requirement Not Met):**
```json
{
  "success": false,
  "message": "Produk Insight 2 memerlukan VIP level 2. Level VIP Anda saat ini: 1"
}
```

**Response (Purchase Limit Reached):**
```json
{
  "success": false,
  "message": "Anda telah mencapai batas pembelian untuk produk Insight 1 (maksimal 1x)"
}
```

---

### 3. GET /api/users/investment/active
**Description:** Get user's active investments grouped by category

**Request:**
```http
GET /api/users/investment/active HTTP/1.1
Host: api.example.com
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "Monitor": [
      {
        "id": 123,
        "user_id": 1,
        "product_id": 1,
        "product_name": "Monitor 1",
        "category": "Monitor",
        "amount": 50000,
        "duration": 70,
        "daily_profit": 15000,
        "total_paid": 35,
        "total_returned": 525000,
        "last_return_at": "2025-10-10T00:00:00Z",
        "next_return_at": "2025-10-11T00:00:00Z",
        "order_id": "INV1234567890",
        "status": "Running"
      }
    ],
    "Insight": [],
    "AutoPilot": []
  }
}
```

**Changes:**
- Products now grouped by category instead of product name
- Added `category` and `product_name` fields
- Removed `percentage` field

---

### 4. GET /api/users/investments
**Description:** List all user investments with pagination

**Response Changes:**
```json
{
  "success": true,
  "message": "Successfully",
  "data": {
    "investments": [
      {
        "id": 123,
        "user_id": 1,
        "product_id": 1,
        "category": "Monitor",
        "amount": 50000,
        "daily_profit": 15000,
        "duration": 70,
        "total_paid": 70,
        "total_returned": 1050000,
        "order_id": "INV1234567890",
        "status": "Completed",
        "created_at": "2025-08-01T00:00:00Z",
        "updated_at": "2025-10-10T00:00:00Z"
      }
    ]
  }
}
```

---

### 5. POST /api/payments/kyta/webhook
**Description:** Webhook for payment confirmation (internal changes)

**Internal Changes:**
- Updates `total_monitor_invest` for Monitor category investments
- Calculates and updates VIP level automatically
- VIP level update occurs only for Monitor category

**No API interface changes** - this is called by the payment gateway

---

### 6. POST /api/cron/daily-returns
**Description:** Cron job for daily profit distribution (internal changes)

**Internal Changes:**

**Monitor Category:**
- Daily profit accumulates in `total_returned`
- User balance is NOT updated daily
- On completion day:
  - Total profit = `daily_profit × duration` is paid to balance
  - Transaction created with type "return"
  - Team bonuses calculated on total profit

**Insight/AutoPilot Categories:**
- Daily profit immediately paid to balance
- Transaction created daily
- Team bonuses calculated daily

**No API interface changes** - this is triggered by cron

---

## User Model Changes

### New Field: `total_monitor_invest`

```go
type User struct {
    // ... existing fields
    TotalMonitorInvest float64 `gorm:"column:total_monitor_invest;type:decimal(15,2);default:0" json:"total_monitor_invest"`
    // ... existing fields
}
```

**Purpose:** Track cumulative Monitor category investments for VIP level calculation

**Updated by:** 
- Payment webhook when Monitor investment is confirmed
- Only Monitor category investments increment this value

---

## Investment Flow Changes

### OLD Flow:
1. User selects product
2. User enters amount (within min/max range)
3. System validates amount range
4. Creates investment with user-specified amount
5. Daily profit calculated: `(amount × percentage / 100 + amount) / duration`

### NEW Flow:

**For Monitor Products:**
1. User selects Monitor product (no amount input)
2. System uses product's fixed amount
3. Creates investment with product's fixed amount and daily_profit
4. Payment confirmed → Updates `total_invest` and `total_monitor_invest`
5. VIP level recalculated based on `total_monitor_invest`
6. Daily: Profit accumulates in `total_returned`, balance unchanged
7. On completion: Total profit paid to balance in single transaction

**For Insight/AutoPilot Products:**
1. User selects product (no amount input)
2. System checks VIP requirement
3. If VIP insufficient → Reject with error message
4. If VIP sufficient → Creates investment with product's fixed amount
5. Payment confirmed → Updates `total_invest` only (not `total_monitor_invest`)
6. On completion (1 day): Profit paid to balance

---

## Bonus System Changes ⭐ MAJOR SIMPLIFICATION

### OLD System (Removed):
- **Team Management Bonuses**: 5%, 2%, 1% for levels 1-3 paid on daily profits
- **Investor Recommendation Bonuses**: 10% level 1, 5% level 2, 1% level 3
- Complex hierarchy with level checking
- Multiple transactions for each profit distribution

### NEW System (Simplified):
**Only ONE bonus type remains:**

#### Investor Recommendation Bonus
- **30% to direct referrer (level 1 only)**
- Paid once when investment payment is confirmed
- Calculated on investment amount (not profit)
- No level 2 or level 3 bonuses
- No team management bonuses

**Example:**
- User A invests 1,000,000 in Monitor 1
- User A's direct referrer (User B) gets 300,000 immediately
- No bonuses for User B's referrer or anyone else

**Benefits:**
- Much simpler calculation
- Easier to understand for users
- Less database transactions
- No complex team hierarchy logic
- Better for system performance

---

## Migration Guide

### For Existing Users:
1. **Existing investments** will continue with old logic (open amount system)
2. **New investments** must use new fixed-product system
3. Users can see their `total_monitor_invest` to check VIP eligibility

### Database Migration:
```sql
-- Create categories table
CREATE TABLE `categories` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(100) NOT NULL,
  `description` text,
  `profit_type` enum('locked','unlocked') NOT NULL DEFAULT 'unlocked',
  `status` enum('Active','Inactive') NOT NULL DEFAULT 'Active',
  PRIMARY KEY (`id`)
);

-- Insert default categories
INSERT INTO `categories` (id, name, description, profit_type, status) VALUES
(1, 'Monitor', 'Profit terkunci', 'locked', 'Active'),
(2, 'Insight', 'Profit langsung', 'unlocked', 'Active'),
(3, 'AutoPilot', 'Profit langsung', 'unlocked', 'Active');

-- Add new column to users table
ALTER TABLE users ADD COLUMN total_invest_vip decimal(15,2) DEFAULT '0.00' 
  COMMENT 'Total locked category investments for VIP level calculation';

-- Modify products table (backup first!)
ALTER TABLE products 
  DROP COLUMN minimum,
  DROP COLUMN maximum,
  DROP COLUMN percentage,
  ADD COLUMN category_id int UNSIGNED NOT NULL AFTER id,
  ADD COLUMN purchase_limit int DEFAULT '0' COMMENT 'Max purchases per user',
  ADD FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;

-- Modify investments table
ALTER TABLE investments
  DROP COLUMN percentage,
  ADD COLUMN category_id int UNSIGNED NOT NULL AFTER product_id,
  ADD FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
```

### Code Migration:
All controllers and models have been updated. Key files changed:
- ✅ `database/db.sql` - Complete schema with categories table
- ✅ `models/category.go` - NEW model for categories
- ✅ `models/product.go` - Uses category_id, removed percentage
- ✅ `models/investment.go` - Uses category_id
- ✅ `models/user.go` - Added total_monitor_invest
- ✅ `controllers/admins/categories.go` - NEW admin API for categories
- ✅ `controllers/admins/products.go` - NEW admin API for products (POST/PUT/DELETE)
- ✅ `controllers/products.go` - Dynamic category-based grouping
- ✅ `controllers/users/investment.go` - Complete refactor with simplified bonuses

---

## Testing Checklist

### Admin Categories API:
- [ ] GET /api/admin/categories lists all categories
- [ ] POST /api/admin/categories creates new category
- [ ] PUT /api/admin/categories/{id} updates category name
- [ ] DELETE /api/admin/categories/{id} prevented if products exist
- [ ] Category profit_type determines profit behavior

### Admin Products API:
- [ ] GET /api/admin/products lists all with category info
- [ ] POST /api/admin/products creates product with category_id
- [ ] PUT /api/admin/products/{id} can change category
- [ ] DELETE /api/admin/products/{id} prevented if investments exist
- [ ] Validation works (amount > 0, duration > 0, etc.)

### User Product API:
- [ ] GET /api/products returns products grouped by dynamic category names
- [ ] Categories display correctly even if renamed
- [ ] Products show category relationship

### Investment Creation:
- [ ] Products use fixed amount (no amount parameter)
- [ ] VIP requirement validation works
- [ ] Purchase limit validation works:
  - [ ] Insight products: Can only buy 1x per product
  - [ ] AutoPilot 1 & 2: Can buy 2x per product
  - [ ] AutoPilot 3 & 4: Can only buy 1x per product
  - [ ] Monitor: Unlimited purchases
  - [ ] Error message shows when limit reached
- [ ] category_id saved correctly in investments

### VIP Level System:
- [ ] VIP level updates only for "locked" profit_type categories
- [ ] VIP level calculated correctly at thresholds (50k, 1.2M, 7M, 30M, 150M)
- [ ] "unlocked" categories don't affect VIP level
- [ ] total_monitor_invest updated correctly

### Profit Distribution:
- [ ] Locked profit: Accumulates, paid at completion
- [ ] Unlocked profit: Paid immediately on completion
- [ ] Cron job checks profit_type from database

### Bonus System:
- [ ] ✅ Only direct referrer gets bonus (level 1)
- [ ] ✅ Bonus is 30% of investment amount
- [ ] ✅ Bonus paid once at payment confirmation
- [ ] ✅ No level 2 or level 3 bonuses
- [ ] ✅ No team management bonuses
- [ ] Spin ticket still given for >= 100k investments

### User Balance:
- [ ] Locked category: Balance not updated daily
- [ ] Locked category: Completion pays total profit at once
- [ ] Unlocked category: Profit paid on completion

---

## Important Notes

1. **Dynamic Categories:** Categories are database-driven. Admin can rename categories without code changes. Always fetch from database, never hardcode category names.

2. **Profit Type:** The `profit_type` field in categories table determines profit behavior:
   - `locked`: Profit accumulates, paid at completion
   - `unlocked`: Profit paid immediately on completion

3. **VIP Level Calculation:** Only investments in categories with `profit_type='locked'` count toward VIP level (stored in `total_invest_vip`). VIP level stored in `users.level` (0-5).

4. **Foreign Keys:** Products and Investments are protected by foreign keys. Cannot delete categories or products that are in use.

5. **Simplified Bonuses:** Only ONE bonus type: 30% direct referral. No team bonuses, no multi-level. Much simpler!

6. **Category Relations:** Always use `Preload("Category")` when querying products or investments to get category info.

7. **Admin Flexibility:** Admin can:
   - Create new categories with custom names
   - Add products to any category
   - Change product categories
   - Rename categories (users will see updated names)

8. **Purchase Limits:** Products have `purchase_limit` field:
   - `0` = Unlimited (Monitor products)
   - `1` = Once per lifetime (All Insight, AutoPilot 3 & 4)
   - `2` = Twice per lifetime (AutoPilot 1 & 2)
   - System counts ONLY **Running and Completed** investments (paid only)
   - Pending (unpaid) and Cancelled do NOT count toward limit

9. **Investment Tracking:**
   - `total_invest`: ALL investments (shown to user as "Total Investasi")
   - `total_invest_vip`: Only locked categories (for VIP calculation)
   - `level`: VIP level (0-5), auto-calculated from `total_invest_vip`
   - All updated when payment is confirmed

10. **VIP Validation:**
    - Uses `users.level` field for validation
    - Purchase allowed if: `user.level >= product.required_vip`
    - Frontend shows lock icon and disables button if not eligible

11. **Testing Edge Cases:**
    - VIP thresholds: 49,999 vs 50,000
    - Purchase limit: User trying to buy after limit reached
    - Category deletion with existing products
    - Product deletion with existing investments
    - Profit distribution on exact completion day
    - Monitor vs Insight investment impact on VIP

---

---

## Final Changes Summary (v2.1)

### Latest Updates:
1. ✅ Field renamed: `total_monitor_invest` → `total_invest_vip` (more clear naming)
2. ✅ VIP validation simplified: Uses `users.level` directly
3. ✅ Frontend: Buy button per product (not single button)
4. ✅ Frontend: VIP status card in dashboard
5. ✅ Frontend: New `/vip` page for detailed VIP info
6. ✅ Admin: Category and product management routes added
7. ✅ Admin: Investment API returns category info

### Files Updated in v2.1:
- `database/db.sql` - Field name update
- `models/user.go` - TotalInvestVIP field
- `controllers/users/investment.go` - Field references updated
- `pages/dashboard.js` - VIP card, buy buttons per product
- `pages/vip.js` - NEW dedicated VIP page
- `pages/portofolio.js` - Category badges
- `components/InvestmentModal.js` - Category warnings
- `routes/admins.go` - New routes

---

## Contact & Support

For questions about this update, please contact the development team.

**Version:** 2.1  
**Date:** October 12, 2025  
**Author:** CIROOS Development Team

