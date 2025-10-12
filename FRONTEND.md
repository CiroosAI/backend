# Frontend Implementation Guide - CIROOS V2

## Overview
Panduan lengkap perubahan tampilan dan fitur frontend untuk mengakomodasi sistem investasi baru dengan kategori dinamis, purchase limit, dan VIP system yang telah direvisi.

---

## Table of Contents
1. [User Pages Changes](#user-pages-changes)
2. [Admin Pages (New)](#admin-pages-new)
3. [Components & UI Elements](#components--ui-elements)
4. [API Integration](#api-integration)
5. [Display Logic](#display-logic)

---

## User Pages Changes

### 1. Product List Page (Halaman Produk)

#### OLD Display:
```
[Bintang 1]
Min: Rp 30.000 - Max: Rp 1.000.000
Profit: 100%
Duration: 200 hari
[Input amount field]
```

#### NEW Display:
```
╔═══════════════ MONITOR ═══════════════╗
║ Monitor 1                              ║
║ Investasi: Rp 50.000                   ║
║ Profit Harian: Rp 15.000               ║
║ Durasi: 70 hari                        ║
║ Total Return: Rp 1.050.000             ║
║ VIP: Tidak perlu                       ║
║ Limit: Unlimited                       ║
║ [BELI SEKARANG]                        ║
╚════════════════════════════════════════╝

╔═══════════════ INSIGHT ════════════════╗
║ Insight 1                              ║
║ Investasi: Rp 50.000                   ║
║ Profit: Rp 20.000                      ║
║ Durasi: 1 hari                         ║
║ Total Return: Rp 70.000                ║
║ VIP: Level 1 Required 🔒               ║
║ Limit: 1x pembelian                   ║
║ Status: ✅ Tersedia / ❌ Habis         ║
║ [BELI SEKARANG] / [SOLD OUT]          ║
╚════════════════════════════════════════╝

╔═══════════════ AUTOPILOT ══════════════╗
║ AutoPilot 1                            ║
║ Investasi: Rp 80.000                   ║
║ Profit: Rp 70.000                      ║
║ Durasi: 1 hari                         ║
║ Total Return: Rp 150.000               ║
║ VIP: Level 3 Required 🔒               ║
║ Limit: 2x pembelian (1/2 tersisa)     ║
║ [BELI SEKARANG]                        ║
╚════════════════════════════════════════╝
```

#### Key Changes:
- ✅ **Group products by category name** (dari API response)
- ✅ **Remove amount input field** (fixed amount dari API)
- ✅ **Show purchase limit** dengan progress (e.g., "1/2 tersisa")
- ✅ **Show VIP requirement** dengan icon lock jika belum memenuhi
- ✅ **Calculate & show total return** = amount + (daily_profit × duration)
- ✅ **Disable button** jika:
  - User VIP level kurang
  - Purchase limit sudah tercapai
  - Product inactive

#### Responsive Behavior:
```javascript
// Pseudo-code untuk button state
function getButtonState(product, userVIPLevel, userPurchaseCount) {
  if (product.status !== 'Active') {
    return { disabled: true, text: 'Tidak Tersedia' };
  }
  
  if (product.required_vip > userVIPLevel) {
    return { disabled: true, text: `Butuh VIP ${product.required_vip}` };
  }
  
  if (product.purchase_limit > 0 && userPurchaseCount >= product.purchase_limit) {
    return { disabled: true, text: 'Limit Tercapai' };
  }
  
  return { disabled: false, text: 'Beli Sekarang' };
}
```

---

### 2. Investment History Page (Riwayat Investasi)

#### NEW Grouped Display:
```
╔══════════════════════════════════════════╗
║           MONITOR (Profit Terkunci)      ║
╠══════════════════════════════════════════╣
║ Monitor 1 - #INV123456                   ║
║ Investasi: Rp 50.000                     ║
║ Profit Harian: Rp 15.000                 ║
║ Progress: 35/70 hari (50%)               ║
║ [████████░░░░░░░░]                       ║
║ Profit Terkumpul: Rp 525.000            ║
║ ⚠️ Dibayar saat selesai                  ║
║ Estimasi selesai: 12 Nov 2025           ║
║ Status: Running 🟢                       ║
╚══════════════════════════════════════════╝

╔══════════════════════════════════════════╗
║           INSIGHT (Profit Langsung)      ║
╠══════════════════════════════════════════╣
║ Insight 1 - #INV123457                   ║
║ Investasi: Rp 50.000                     ║
║ Profit: Rp 20.000                        ║
║ Total Received: Rp 70.000 ✅             ║
║ Completed: 10 Okt 2025                   ║
║ Status: Completed 🎉                     ║
╚══════════════════════════════════════════╝
```

#### Key Features:
- ✅ **Group by category** (Monitor / Insight / AutoPilot)
- ✅ **Show different info for locked vs unlocked**:
  - Locked (Monitor): Show accumulated profit + warning
  - Unlocked (Insight/AutoPilot): Show completed profit
- ✅ **Progress bar** untuk Monitor
- ✅ **Status badges** dengan warna berbeda
- ✅ **Empty state** untuk kategori tanpa investasi

---

### 3. Dashboard / Profile Page

#### NEW VIP Level Display:
```
╔══════════════════════════════════════════╗
║              PROFIL ANDA                 ║
╠══════════════════════════════════════════╣
║ Nama: John Doe                           ║
║ VIP Level: 2 ⭐⭐                         ║
║                                          ║
║ Progress ke VIP 3:                       ║
║ Rp 800.000 / Rp 7.000.000               ║
║ [████░░░░░░░░░░░░░░░] 11%               ║
║                                          ║
║ Investasi Monitor: Rp 800.000           ║
║ Total Investasi: Rp 1.200.000           ║
║                                          ║
║ ℹ️ Hanya investasi Monitor yang          ║
║    menaikkan level VIP                   ║
╚══════════════════════════════════════════╝

╔══════════════════════════════════════════╗
║           STATUS PRODUK                  ║
╠══════════════════════════════════════════╣
║ Monitor: ✅ Semua tersedia               ║
║ Insight 1: ✅ Tersedia                   ║
║ Insight 2: ❌ Butuh VIP 2                ║
║ AutoPilot: ❌ Butuh VIP 3                ║
╚══════════════════════════════════════════╝
```

#### Key Changes:
- ✅ **Show VIP level** dengan visual stars/badges
- ✅ **Progress bar to next VIP** berdasarkan `total_monitor_invest`
- ✅ **Separate display**:
  - "Investasi Monitor" → `total_monitor_invest`
  - "Total Investasi" → `total_invest`
- ✅ **Info tooltip**: Explain kenapa ada 2 angka berbeda
- ✅ **Product availability summary** per kategori

---

### 4. Purchase Confirmation Modal

#### NEW Confirmation Display:
```
╔══════════════════════════════════════════╗
║       KONFIRMASI PEMBELIAN               ║
╠══════════════════════════════════════════╣
║ Produk: Insight 1                        ║
║ Kategori: Insight (Profit Langsung)      ║
║                                          ║
║ Investasi: Rp 50.000                     ║
║ Profit Harian: Rp 20.000                 ║
║ Durasi: 1 hari                           ║
║ ─────────────────────────────────────    ║
║ Total Return: Rp 70.000                  ║
║                                          ║
║ ⚠️ PERHATIAN:                            ║
║ • Produk ini LIMITED 1x pembelian        ║
║ • Profit dibayar langsung saat selesai   ║
║ • Tidak menambah VIP level               ║
║                                          ║
║ [BATAL]  [LANJUTKAN PEMBAYARAN]         ║
╚══════════════════════════════════════════╝
```

#### Different Warnings per Category:
```javascript
const warnings = {
  Monitor: [
    'Profit dikumpulkan dan dibayar saat investasi selesai',
    'Investasi ini akan menambah VIP level Anda',
    'Bisa dibeli berkali-kali'
  ],
  Insight: [
    'LIMITED: Hanya bisa dibeli 1x selamanya',
    'Profit dibayar langsung saat selesai',
    'TIDAK menambah VIP level'
  ],
  AutoPilot: [
    'LIMITED: Bisa dibeli maksimal 1-2x selamanya',
    'Profit dibayar langsung saat selesai',
    'TIDAK menambah VIP level',
    'Memerlukan VIP level 3'
  ]
}
```

---

### 5. Error Messages Display

#### User-Friendly Error Messages:
```javascript
const errorMessages = {
  'vip_required': {
    icon: '🔒',
    title: 'VIP Level Tidak Cukup',
    message: 'Produk {productName} memerlukan VIP level {requiredVIP}.',
    action: 'Tingkatkan VIP dengan investasi Monitor',
    showVIPProgress: true
  },
  
  'purchase_limit_reached': {
    icon: '⛔',
    title: 'Batas Pembelian Tercapai',
    message: 'Anda sudah membeli {productName} sebanyak {limit}x.',
    action: 'Coba produk lain',
    showAlternatives: true
  },
  
  'insufficient_balance': {
    icon: '💰',
    title: 'Saldo Tidak Cukup',
    message: 'Saldo Anda: Rp {balance}. Dibutuhkan: Rp {required}',
    action: 'Top up atau pilih produk lain'
  }
}
```

---

## Admin Pages (NEW)

### 1. Categories Management Page (NEW)

#### Layout:
```
╔════════════════════════════════════════════════════════╗
║              KELOLA KATEGORI PRODUK                    ║
╠════════════════════════════════════════════════════════╣
║ [+ Tambah Kategori Baru]                    [Search]   ║
╠════════════════════════════════════════════════════════╣
║ ID │ Nama      │ Profit Type │ Produk │ Status │ Aksi ║
╠════════════════════════════════════════════════════════╣
║ 1  │ Monitor   │ Locked      │ 7      │ 🟢    │ ✏️ 🗑️║
║ 2  │ Insight   │ Unlocked    │ 5      │ 🟢    │ ✏️ 🗑️║
║ 3  │ AutoPilot │ Unlocked    │ 4      │ 🟢    │ ✏️ 🗑️║
╚════════════════════════════════════════════════════════╝
```

#### Add/Edit Category Form:
```
╔══════════════════════════════════════════╗
║       TAMBAH/EDIT KATEGORI               ║
╠══════════════════════════════════════════╣
║ Nama Kategori: *                         ║
║ [________________________]               ║
║                                          ║
║ Deskripsi:                               ║
║ [________________________]               ║
║ [________________________]               ║
║                                          ║
║ Tipe Profit: *                           ║
║ ( ) Locked - Dibayar saat selesai        ║
║ (•) Unlocked - Dibayar langsung          ║
║                                          ║
║ Status:                                  ║
║ [✓] Active  [ ] Inactive                 ║
║                                          ║
║ ℹ️ Kategori dengan "Locked" akan         ║
║    menambah VIP level user               ║
║                                          ║
║ [BATAL]  [SIMPAN]                        ║
╚══════════════════════════════════════════╝
```

#### Features:
- ✅ List all categories dengan jumlah produk
- ✅ Add new category dengan validation
- ✅ Edit category name (users akan lihat nama baru)
- ✅ Delete dengan protection (tidak bisa hapus jika ada produk)
- ✅ Toggle active/inactive

---

### 2. Products Management Page (UPDATED)

#### Layout:
```
╔══════════════════════════════════════════════════════════════╗
║                  KELOLA PRODUK INVESTASI                     ║
╠══════════════════════════════════════════════════════════════╣
║ [+ Tambah Produk Baru]           [Filter: Semua ▼] [Search] ║
╠══════════════════════════════════════════════════════════════╣
║ ID │ Nama       │ Kategori │ Amount    │ Profit   │ Limit  ║
╠══════════════════════════════════════════════════════════════╣
║ 1  │ Monitor 1  │ Monitor  │ 50.000    │ 15.000   │ ∞     ║
║ 8  │ Insight 1  │ Insight  │ 50.000    │ 20.000   │ 1x    ║
║ 13 │ AutoPilot 1│ AutoPilot│ 80.000    │ 70.000   │ 2x    ║
║    │            │          │           │          │ ✏️ 🗑️ ║
╚══════════════════════════════════════════════════════════════╝
```

#### Add/Edit Product Form:
```
╔══════════════════════════════════════════╗
║         TAMBAH/EDIT PRODUK               ║
╠══════════════════════════════════════════╣
║ Kategori: *                              ║
║ [Monitor        ▼]                       ║
║                                          ║
║ Nama Produk: *                           ║
║ [Monitor 8________________]              ║
║                                          ║
║ Jumlah Investasi (Rp): *                 ║
║ [50.000.000_______________]              ║
║                                          ║
║ Profit Harian (Rp): *                    ║
║ [20.000.000_______________]              ║
║                                          ║
║ Durasi (hari): *                         ║
║ [45_______]                              ║
║                                          ║
║ VIP Required:                            ║
║ [0_______] (0 = tidak perlu)             ║
║                                          ║
║ Purchase Limit:                          ║
║ [0_______] (0 = unlimited)               ║
║ ℹ️ 1 = sekali, 2 = dua kali              ║
║                                          ║
║ Status:                                  ║
║ [✓] Active  [ ] Inactive                 ║
║                                          ║
║ ─────── PREVIEW ──────                   ║
║ Total Return: Rp 920.000.000             ║
║ (Amount + Profit × Duration)             ║
║                                          ║
║ [BATAL]  [SIMPAN]                        ║
╚══════════════════════════════════════════╝
```

#### Features:
- ✅ Dropdown kategori (dari API categories)
- ✅ Auto-calculate total return untuk preview
- ✅ Validation:
  - Amount > 0
  - Daily profit > 0
  - Duration > 0
  - Purchase limit >= 0
- ✅ Delete dengan protection (tidak bisa hapus jika ada investasi)

---

## Components & UI Elements

### 1. VIP Level Badge Component

```jsx
<VIPBadge level={userLevel}>
  VIP {userLevel} {"⭐".repeat(userLevel)}
</VIPBadge>

// Color scheme:
// VIP 0: Gray
// VIP 1: Bronze
// VIP 2: Silver
// VIP 3: Gold
// VIP 4: Platinum
// VIP 5: Diamond
```

### 2. Purchase Limit Indicator

```jsx
<PurchaseLimitBadge 
  limit={product.purchase_limit}
  used={userPurchaseCount}
>
  {limit === 0 ? "∞ Unlimited" : `${used}/${limit} digunakan`}
</PurchaseLimitBadge>
```

### 3. Category Badge

```jsx
<CategoryBadge 
  category={product.category}
  profitType={product.category.profit_type}
>
  {category.name}
  {profitType === 'locked' ? '🔒' : '⚡'}
</CategoryBadge>
```

### 4. Investment Status Badge

```jsx
const statusConfig = {
  Pending: { color: 'yellow', icon: '⏳', text: 'Menunggu Pembayaran' },
  Running: { color: 'green', icon: '🟢', text: 'Berjalan' },
  Completed: { color: 'blue', icon: '✅', text: 'Selesai' },
  Cancelled: { color: 'red', icon: '❌', text: 'Dibatalkan' },
  Suspended: { color: 'orange', icon: '⏸️', text: 'Ditangguhkan' }
}
```

### 5. Profit Type Indicator

```jsx
<ProfitTypeIndicator type={category.profit_type}>
  {type === 'locked' 
    ? '🔒 Profit Terkunci (dibayar saat selesai)'
    : '⚡ Profit Langsung (dibayar setelah durasi)'
  }
</ProfitTypeIndicator>
```

---

## API Integration

### 1. Get Products (User)

**Endpoint:** `GET /api/products`

**Response Structure:**
```json
{
  "success": true,
  "data": {
    "Monitor": [
      {
        "id": 1,
        "category_id": 1,
        "name": "Monitor 1",
        "amount": 50000,
        "daily_profit": 15000,
        "duration": 70,
        "required_vip": 0,
        "purchase_limit": 0,
        "status": "Active",
        "category": {
          "id": 1,
          "name": "Monitor",
          "profit_type": "locked"
        }
      }
    ],
    "Insight": [...],
    "AutoPilot": [...]
  }
}
```

**Frontend Usage:**
```javascript
// Fetch products
const response = await fetch('/api/products');
const data = await response.json();

// Display grouped by category
Object.entries(data.data).forEach(([categoryName, products]) => {
  renderCategorySection(categoryName, products);
});

// Calculate display values
products.forEach(product => {
  product.totalReturn = product.amount + (product.daily_profit * product.duration);
  product.isUnlimited = product.purchase_limit === 0;
});
```

### 2. Create Investment (User)

**Endpoint:** `POST /api/users/investments`

**Request:**
```json
{
  "product_id": 8,
  "payment_method": "QRIS",
  "payment_channel": ""
}
```

**Error Responses to Handle:**
```javascript
const handleInvestmentError = (error) => {
  if (error.message.includes('VIP level')) {
    showVIPRequiredModal(productName, requiredVIP, currentVIP);
  } else if (error.message.includes('batas pembelian')) {
    showPurchaseLimitModal(productName, limit);
  } else if (error.message.includes('Produk tidak ditemukan')) {
    showProductUnavailableModal();
  }
}
```

### 3. Get User Investments

**Endpoint:** `GET /api/users/investment/active`

**Response Structure:**
```json
{
  "success": true,
  "data": {
    "Monitor": [
      {
        "id": 123,
        "product_name": "Monitor 1",
        "category_name": "Monitor",
        "amount": 50000,
        "daily_profit": 15000,
        "duration": 70,
        "total_paid": 35,
        "total_returned": 525000,
        "status": "Running"
      }
    ],
    "Insight": [],
    "AutoPilot": []
  }
}
```

### 4. Admin APIs

#### Get Categories (Admin)
```javascript
GET /api/admin/categories
→ List all categories untuk dropdown dan table
```

#### Create Product (Admin)
```javascript
POST /api/admin/products
Body: {
  category_id, name, amount, daily_profit, 
  duration, required_vip, purchase_limit, status
}
```

---

## Display Logic

### 1. Calculate Purchase Count

```javascript
async function getUserPurchaseCount(userId, productId) {
  // Client-side calculation atau dari API
  const investments = await fetchUserInvestments(userId);
  
  return investments.filter(inv => 
    inv.product_id === productId && 
    ['Pending', 'Running', 'Completed'].includes(inv.status)
  ).length;
}
```

### 2. Calculate VIP Progress

```javascript
const VIP_THRESHOLDS = {
  1: 50000,
  2: 1200000,
  3: 7000000,
  4: 30000000,
  5: 150000000
};

function getVIPProgress(totalMonitorInvest) {
  const currentVIP = Object.entries(VIP_THRESHOLDS)
    .reverse()
    .find(([_, threshold]) => totalMonitorInvest >= threshold)?.[0] || 0;
  
  const nextVIP = parseInt(currentVIP) + 1;
  const nextThreshold = VIP_THRESHOLDS[nextVIP];
  
  if (!nextThreshold) {
    return { current: currentVIP, next: null, progress: 100, remaining: 0 };
  }
  
  const progress = (totalMonitorInvest / nextThreshold) * 100;
  const remaining = nextThreshold - totalMonitorInvest;
  
  return { current: currentVIP, next: nextVIP, progress, remaining };
}
```

### 3. Product Availability Check

```javascript
function isProductAvailable(product, user) {
  // Check status
  if (product.status !== 'Active') return false;
  
  // Check VIP requirement
  if (product.required_vip > user.level) return false;
  
  // Check purchase limit
  if (product.purchase_limit > 0) {
    const userPurchaseCount = getUserPurchaseCount(user.id, product.id);
    if (userPurchaseCount >= product.purchase_limit) return false;
  }
  
  return true;
}
```

### 4. Display Total Return

```javascript
function calculateTotalReturn(product) {
  // Total return = investment amount + total profit
  return product.amount + (product.daily_profit * product.duration);
}

function formatTotalReturn(product) {
  const total = calculateTotalReturn(product);
  return {
    amount: product.amount,
    profit: product.daily_profit * product.duration,
    total: total,
    formatted: `Rp ${total.toLocaleString('id-ID')}`
  };
}
```

---

## Responsive Design Considerations

### Mobile View:
```
╔═══════════════════════════╗
║       MONITOR             ║
╠═══════════════════════════╣
║ [Card: Monitor 1]         ║
║ Rp 50.000                 ║
║ Profit: Rp 15.000/hari    ║
║ Total: Rp 1.050.000       ║
║ [BELI]                    ║
╟───────────────────────────╢
║ [Card: Monitor 2]         ║
║ ...                       ║
╚═══════════════════════════╝
```

### Tablet/Desktop:
```
╔════════════════════════════════════════════════════╗
║                    MONITOR                         ║
╠════════════════════════════════════════════════════╣
║ [Card] [Card] [Card] [Card]                        ║
║ Monitor Monitor Monitor Monitor                    ║
║   1      2       3      4                          ║
╚════════════════════════════════════════════════════╝
```

---

## Loading States

### Product Loading:
```jsx
<Skeleton>
  <SkeletonCategory />
  <SkeletonProductCard count={3} />
</Skeleton>
```

### Investment Status Loading:
```jsx
<LoadingSpinner text="Memproses investasi..." />
```

---

## Empty States

### No Products in Category:
```
╔══════════════════════════════════╗
║        📦                        ║
║   Belum ada produk               ║
║   di kategori ini                ║
╚══════════════════════════════════╝
```

### No Investments:
```
╔══════════════════════════════════╗
║        💼                        ║
║   Belum ada investasi            ║
║   di kategori ini                ║
║                                  ║
║   [LIHAT PRODUK]                 ║
╚══════════════════════════════════╝
```

---

## Notifications & Alerts

### Success Messages:
```javascript
notifications = {
  investmentCreated: {
    type: 'success',
    title: 'Investasi Berhasil Dibuat',
    message: 'Silakan lakukan pembayaran dalam 15 menit',
    action: 'Lihat Detail Pembayaran'
  },
  
  investmentCompleted: {
    type: 'success',
    title: 'Investasi Selesai! 🎉',
    message: 'Profit Rp {amount} telah masuk ke saldo Anda',
    action: 'Lihat Saldo'
  },
  
  vipLevelUp: {
    type: 'celebration',
    title: 'Selamat! VIP Level Naik! 🌟',
    message: 'Anda naik ke VIP Level {newLevel}',
    action: 'Lihat Produk Baru'
  }
}
```

### Warning Messages:
```javascript
warnings = {
  limitedProduct: {
    type: 'warning',
    icon: '⚠️',
    message: 'Produk ini LIMITED! Hanya bisa dibeli {limit}x selamanya'
  },
  
  lockedProfit: {
    type: 'info',
    icon: 'ℹ️',
    message: 'Profit akan dibayar setelah investasi selesai'
  }
}
```

---

## Summary Checklist

### User Pages:
- [ ] Product list dengan grouping dinamis per kategori
- [ ] Purchase limit indicator (unlimited / 1x / 2x)
- [ ] VIP requirement badge & validation
- [ ] Total return calculation display
- [ ] Investment history grouped by category
- [ ] Different display for locked vs unlocked profit
- [ ] VIP progress bar dengan 2 angka berbeda
- [ ] Purchase confirmation modal dengan warnings
- [ ] Error handling yang user-friendly

### Admin Pages:
- [ ] Categories management (CRUD)
- [ ] Products management dengan category dropdown
- [ ] Purchase limit field di form produk
- [ ] Total return preview calculator
- [ ] Delete protection warnings

### Components:
- [ ] VIP badge dengan gradasi warna
- [ ] Purchase limit badge
- [ ] Category badge dengan icon
- [ ] Status badges
- [ ] Profit type indicator
- [ ] Progress bars

### Mobile Responsiveness:
- [ ] Card layout untuk mobile
- [ ] Collapsible categories
- [ ] Touch-friendly buttons
- [ ] Optimized forms

---

## Notes for Developers

1. **Dynamic Categories**: NEVER hardcode category names. Always fetch from API.

2. **Two Investment Fields**: 
   - Display `total_invest` to user as "Total Investasi"
   - Use `total_monitor_invest` only for VIP calculation

3. **Purchase Limit**: 
   - Count Pending + Running + Completed status
   - Don't count Cancelled or Suspended

4. **Profit Calculation**:
   - Locked: Show "Terkumpul" dengan warning
   - Unlocked: Show as normal completed payment

5. **Color Coding**:
   - Monitor (Locked): Blue/Purple theme
   - Insight (Unlocked): Green theme  
   - AutoPilot (Unlocked): Orange theme

6. **Icons Recommendation**:
   - Monitor: 🔒 or 📊
   - Insight: ⚡ or 💡
   - AutoPilot: 🚀 or 🤖
   - VIP: ⭐ or 👑
   - Unlimited: ∞
   - Limited: ⚠️ or 🔢

---

## Contact & Support

Untuk pertanyaan implementasi frontend, hubungi team development.

**Version:** 2.0  
**Last Updated:** October 12, 2025

