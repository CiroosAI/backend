-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Waktu pembuatan: 26 Sep 2025 pada 12.14
-- Versi server: 8.4.3
-- Versi PHP: 8.3.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `sf`
--

-- --------------------------------------------------------

--
-- Struktur dari tabel `admins`
--

CREATE TABLE `admins` (
  `id` bigint UNSIGNED NOT NULL,
  `username` varchar(191) NOT NULL,
  `password` longtext NOT NULL,
  `name` longtext NOT NULL,
  `email` varchar(191) DEFAULT NULL,
  `role` varchar(191) DEFAULT 'admin',
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Dumping data untuk tabel `admins`
--

INSERT INTO `admins` (`id`, `username`, `password`, `name`, `email`, `role`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'admin', '$2y$10$I4qWolurBpmNKJlQUqb6CeBASh/8Sv59gWu6Ys.m9UsXPLdRLm0du', 'Admin', 'admin@vladevs.com', 'admin', 1, '2000-01-01 00:00:00.000', '2000-01-01 00:00:00.000');

-- --------------------------------------------------------

--
-- Struktur dari tabel `banks`
--

CREATE TABLE `banks` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Bank Rakyat Indonesia, Bank Central Asia, Dana, GoPay',
  `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'BRI, BCA, DANA, GOPAY for payment gateway API',
  `status` enum('Active','Maintenance','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Available banks and e-wallets for withdrawal';

--
-- Dumping data untuk tabel `banks`
--

INSERT INTO `banks` (`id`, `name`, `code`, `status`) VALUES
(1, 'Bank Rakyat Indonesia', 'BRI', 'Active'),
(2, 'Bank Central Asia', 'BCA', 'Active'),
(3, 'Bank Negara Indonesia', 'BNI', 'Active'),
(4, 'Bank Mandiri', 'MANDIRI', 'Active'),
(5, 'Bank Permata', 'PERMATA', 'Active'),
(6, 'Bank CIMB Niaga', 'BNC', 'Active'),
(7, 'Dana', 'DANA', 'Active'),
(8, 'GoPay', 'GOPAY', 'Active'),
(9, 'OVO', 'OVO', 'Active'),
(10, 'ShopeePay', 'SHOPEEPAY', 'Active');

-- --------------------------------------------------------

--
-- Struktur dari tabel `bank_accounts`
--

CREATE TABLE `bank_accounts` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `bank_id` int UNSIGNED NOT NULL,
  `account_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Nama penerima/pemilik rekening',
  `account_number` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Nomor rekening atau nomor e-wallet'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User linked bank accounts and e-wallets';

-- --------------------------------------------------------

--
-- Struktur dari tabel `forums`
--

CREATE TABLE `forums` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  `reward` decimal(15,2) DEFAULT '0.00',
  `description` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `image` varchar(255) NOT NULL,
  `status` enum('Accepted','Pending','Rejected') DEFAULT 'Pending',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `investments`
--

CREATE TABLE `investments` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `product_id` int UNSIGNED NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `percentage` decimal(5,2) NOT NULL,
  `duration` int NOT NULL,
  `daily_profit` decimal(15,2) NOT NULL,
  `total_paid` int NOT NULL DEFAULT '0',
  `total_returned` decimal(15,2) NOT NULL DEFAULT '0.00',
  `last_return_at` datetime DEFAULT NULL,
  `next_return_at` datetime DEFAULT NULL,
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('Pending','Running','Completed','Suspended','Cancelled') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `payments`
--

CREATE TABLE `payments` (
  `id` bigint UNSIGNED NOT NULL,
  `investment_id` int NOT NULL,
  `reference_id` varchar(191) DEFAULT NULL,
  `order_id` varchar(191) NOT NULL,
  `payment_method` varchar(16) DEFAULT NULL,
  `payment_channel` varchar(16) DEFAULT NULL,
  `payment_code` text,
  `payment_link` text,
  `status` varchar(16) NOT NULL DEFAULT 'Pending',
  `expired_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `payment_settings`
--

CREATE TABLE `payment_settings` (
  `id` bigint UNSIGNED NOT NULL,
  `pakasir_api_key` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pakasir_project` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `deposit_amount` decimal(15,2) NOT NULL DEFAULT '0.00',
  `bank_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bank_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `account_number` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `account_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `withdraw_amount` decimal(15,2) NOT NULL DEFAULT '0.00',
  `wishlist_id` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `payment_settings`
--

INSERT INTO `payment_settings` (`id`, `pakasir_api_key`, `pakasir_project`, `deposit_amount`, `bank_name`, `bank_code`, `account_number`, `account_name`, `withdraw_amount`, `wishlist_id`, `created_at`, `updated_at`) VALUES
(1, 'AWD1A2AWD132', 'AWD1SAD2A1W', 10000.00, 'Bank BCA', 'BCA', '1234567890', 'StoneForm Admin', 50000.00, '1', '2025-09-26 12:13:38', '2025-09-26 12:13:38');

-- --------------------------------------------------------

--
-- Struktur dari tabel `products`
--

CREATE TABLE `products` (
  `id` int UNSIGNED NOT NULL,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `minimum` decimal(15,2) NOT NULL,
  `maximum` decimal(15,2) NOT NULL,
  `percentage` decimal(5,2) NOT NULL,
  `duration` int NOT NULL,
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `products`
--

INSERT INTO `products` (`id`, `name`, `minimum`, `maximum`, `percentage`, `duration`, `status`, `created_at`, `updated_at`) VALUES
(1, 'Bintang 1', 30000.00, 1000000.00, 100.00, 200, 'Active', '2025-09-07 02:06:16', '2025-09-17 16:22:49'),
(2, 'Bintang 2', 1500000.00, 3000000.00, 100.00, 67, 'Active', '2025-09-07 02:06:16', '2025-09-10 15:37:28'),
(3, 'Bintang 3', 5000000.00, 10000000.00, 100.00, 40, 'Active', '2025-09-07 02:06:16', '2025-09-10 15:37:32');

-- --------------------------------------------------------

--
-- Struktur dari tabel `refresh_tokens`
--

CREATE TABLE `refresh_tokens` (
  `id` char(64) NOT NULL,
  `user_id` bigint NOT NULL,
  `expires_at` datetime(3) DEFAULT NULL,
  `revoked` tinyint(1) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `revoked_tokens`
--

CREATE TABLE `revoked_tokens` (
  `id` varchar(128) NOT NULL,
  `revoked_at` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `settings`
--

CREATE TABLE `settings` (
  `id` bigint UNSIGNED NOT NULL,
  `name` text NOT NULL,
  `logo` text NOT NULL,
  `min_withdraw` decimal(15,2) NOT NULL,
  `max_withdraw` decimal(15,2) NOT NULL,
  `withdraw_charge` decimal(15,2) NOT NULL,
  `maintenance` BOOLEAN NOT NULL DEFAULT FALSE,
  `closed_register` BOOLEAN NOT NULL DEFAULT FALSE,
  `link_cs` text NOT NULL,
  `link_group` text NOT NULL,
  `link_app` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Dumping data untuk tabel `settings`
--

INSERT INTO `settings` (`id`, `name`, `logo`, `min_withdraw`, `max_withdraw`, `withdraw_charge`, `link_cs`, `link_group`, `link_app`) VALUES
(1, 'Vla Devs', 'logo.png', 33000.00, 10000000.00, 10.00, 'https://t.me/', 'https://t.me/', 'https://vladevs.com');

-- --------------------------------------------------------

--
-- Struktur dari tabel `spin_prizes`
--

CREATE TABLE `spin_prizes` (
  `id` int UNSIGNED NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Unique code untuk validasi claim prize',
  `chance_weight` int NOT NULL COMMENT 'Weight untuk random selection (semakin besar semakin sering muncul)',
  `status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Active',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Available spin wheel prizes';

--
-- Dumping data untuk tabel `spin_prizes`
--

INSERT INTO `spin_prizes` (`id`, `amount`, `code`, `chance_weight`, `status`, `created_at`, `updated_at`) VALUES
(1, 1000.00, 'SPIN_1K', 5000, 'Active', '2025-08-31 02:48:48', '2025-09-18 12:18:21'),
(2, 5000.00, 'SPIN_5K', 500, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:11:12'),
(3, 10000.00, 'SPIN_10K', 300, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:11:16'),
(4, 50000.00, 'SPIN_50K', 30, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:17:32'),
(5, 100000.00, 'SPIN_100K', 10, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:17:28'),
(6, 200000.00, 'SPIN_200K', 5, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:04:43'),
(7, 500000.00, 'SPIN_500K', 2, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:04:46'),
(8, 1000000.00, 'SPIN_1000K', 1, 'Active', '2025-08-31 02:48:48', '2025-09-15 21:50:03');

-- --------------------------------------------------------

--
-- Stand-in struktur untuk tampilan `spin_prizes_with_percentage`
-- (Lihat di bawah untuk tampilan aktual)
--
CREATE TABLE `spin_prizes_with_percentage` (
`id` int unsigned
,`amount` decimal(15,2)
,`code` varchar(20)
,`chance_weight` int
,`chance_percentage` decimal(16,2)
,`status` enum('Active','Inactive')
);

-- --------------------------------------------------------

--
-- Struktur dari tabel `tasks`
--

CREATE TABLE `tasks` (
  `id` int NOT NULL,
  `name` varchar(100) NOT NULL,
  `reward` decimal(15,2) NOT NULL,
  `required_level` int NOT NULL,
  `required_active_members` bigint NOT NULL,
  `status` enum('Active','Inactive') DEFAULT 'Active',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- Dumping data untuk tabel `tasks`
--

INSERT INTO `tasks` (`id`, `name`, `reward`, `required_level`, `required_active_members`, `status`, `created_at`, `updated_at`) VALUES
(1, 'Tugas Perekrutan 1', 15000.00, 1, 5, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(2, 'Tugas Perekrutan 2', 35000.00, 1, 10, 'Active', '2025-09-08 03:57:01', '2025-09-11 22:07:23'),
(3, 'Tugas Perekrutan 3', 200000.00, 1, 50, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(4, 'Tugas Perekrutan 4', 450000.00, 1, 100, 'Active', '2025-09-08 03:57:01', '2025-09-08 03:57:01'),
(5, 'Tugas Perekrutan 5', 1000000.00, 1, 200, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(6, 'Tugas Perekrutan 6', 2750000.00, 1, 500, 'Active', '2025-09-08 03:57:01', '2025-09-08 03:57:01'),
(7, 'Tugas Perekrutan 7', 6000000.00, 1, 1000, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(8, 'Tugas Perekrutan 8', 16000000.00, 1, 2000, 'Active', '2025-09-08 03:57:01', '2025-09-08 04:00:03'),
(9, 'Tugas Perekrutan 9', 35000000.00, 1, 3000, 'Active', '2025-09-08 03:56:19', '2025-09-08 03:56:19'),
(10, 'Tugas Perekrutan 10', 80000000.00, 1, 5000, 'Active', '2025-09-08 03:57:01', '2025-09-08 03:57:01');

-- --------------------------------------------------------

--
-- Struktur dari tabel `transactions`
--

CREATE TABLE `transactions` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `charge` decimal(15,2) NOT NULL DEFAULT '0.00',
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `transaction_flow` enum('debit','credit') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'debit=money out, credit=money in',
  `transaction_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'deposit, withdraw, transfer, refund, bonus, penalty, etc',
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
  `status` enum('Success','Pending','Failed') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User transaction records';

-- --------------------------------------------------------

--
-- Struktur dari tabel `users`
--

CREATE TABLE `users` (
  `id` int NOT NULL,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `number` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `reff_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `reff_by` bigint UNSIGNED DEFAULT NULL,
  `balance` decimal(15,2) DEFAULT '0.00',
  `level` bigint NOT NULL DEFAULT '0',
  `total_invest` decimal(15,2) DEFAULT '0.00',
  `spin_ticket` bigint DEFAULT '0',
  `status` enum('Active','Inactive','Suspend') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'Active',
  `investment_status` enum('Active','Inactive') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'Inactive',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `users`
--

INSERT INTO `users` (`id`, `name`, `number`, `password`, `reff_code`, `reff_by`, `balance`, `level`, `total_invest`, `spin_ticket`, `status`, `investment_status`, `created_at`, `updated_at`) VALUES
(1, 'VLA Users', '8123456789', '$2y$10$fa5X/6ZfpaNZsa07TyzO3ukL/AtxtGLv.6erFIw9KmXFNYyFbE656', 'VLAREFF', 0, 2000.00, 5, 1000.00, 100, 'Active', 'Active', '2025-01-01 00:00:00.000', '2025-01-01 00:00:00.000');

-- --------------------------------------------------------

--
-- Struktur dari tabel `user_spins`
--

CREATE TABLE `user_spins` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `prize_id` int UNSIGNED NOT NULL COMMENT 'Reference to won prize',
  `amount` decimal(15,2) NOT NULL COMMENT 'Amount yang dimenangkan',
  `code` VARCHAR(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Code hadiah yang dimenangkan',
  `won_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User spin wheel history and claims';

-- --------------------------------------------------------

--
-- Struktur dari tabel `user_tasks`
--

CREATE TABLE `user_tasks` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  `task_id` int NOT NULL,
  `claimed_at` datetime DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `withdrawals`
--

CREATE TABLE `withdrawals` (
  `id` int UNSIGNED NOT NULL,
  `user_id` int NOT NULL,
  `bank_account_id` int UNSIGNED NOT NULL COMMENT 'Reference to user linked bank account',
  `amount` decimal(15,2) NOT NULL,
  `charge` decimal(15,2) NOT NULL DEFAULT '0.00',
  `final_amount` decimal(15,2) NOT NULL COMMENT 'amount - charge, calculated amount user receives',
  `order_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('Success','Pending','Failed') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User withdrawal requests';

--
-- Trigger `withdrawals`
--
DELIMITER $$
CREATE TRIGGER `withdrawals_calculate_final_amount` BEFORE INSERT ON `withdrawals` FOR EACH ROW BEGIN
    SET NEW.final_amount = NEW.amount - NEW.charge;
END
$$
DELIMITER ;
DELIMITER $$
CREATE TRIGGER `withdrawals_update_final_amount` BEFORE UPDATE ON `withdrawals` FOR EACH ROW BEGIN
    IF NEW.amount != OLD.amount OR NEW.charge != OLD.charge THEN
        SET NEW.final_amount = NEW.amount - NEW.charge;
    END IF;
END
$$
DELIMITER ;

--
-- Indexes for dumped tables
--

--
-- Indeks untuk tabel `admins`
--
ALTER TABLE `admins`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id` (`id`),
  ADD UNIQUE KEY `uni_admins_username` (`username`),
  ADD UNIQUE KEY `uni_admins_email` (`email`);

--
-- Indeks untuk tabel `banks`
--
ALTER TABLE `banks`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `code` (`code`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_code` (`code`);

--
-- Indeks untuk tabel `bank_accounts`
--
ALTER TABLE `bank_accounts`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_user_account` (`user_id`,`bank_id`,`account_number`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_bank_id` (`bank_id`);

--
-- Indeks untuk tabel `forums`
--
ALTER TABLE `forums`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indeks untuk tabel `investments`
--
ALTER TABLE `investments`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_product_id` (`product_id`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_next_return_at` (`next_return_at`);

--
-- Indeks untuk tabel `payments`
--
ALTER TABLE `payments`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id` (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`);

--
-- Indeks untuk tabel `payment_settings`
--
ALTER TABLE `payment_settings`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`),
  ADD KEY `idx_products_status` (`status`);

--
-- Indeks untuk tabel `refresh_tokens`
--
ALTER TABLE `refresh_tokens`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_refresh_user` (`user_id`),
  ADD KEY `idx_refresh_tokens_user_id` (`user_id`);

--
-- Indeks untuk tabel `revoked_tokens`
--
ALTER TABLE `revoked_tokens`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `settings`
--
ALTER TABLE `settings`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `id` (`id`);

--
-- Indeks untuk tabel `spin_prizes`
--
ALTER TABLE `spin_prizes`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `code` (`code`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_code` (`code`),
  ADD KEY `idx_chance_weight` (`chance_weight`);

--
-- Indeks untuk tabel `tasks`
--
ALTER TABLE `tasks`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `transactions`
--
ALTER TABLE `transactions`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_order_id` (`order_id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_order_id` (`order_id`),
  ADD KEY `idx_transaction_flow` (`transaction_flow`),
  ADD KEY `idx_transaction_type` (`transaction_type`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_created_at` (`created_at`),
  ADD KEY `idx_user_status_created` (`user_id`,`status`,`created_at`),
  ADD KEY `idx_user_type_created` (`user_id`,`transaction_type`,`created_at`);

--
-- Indeks untuk tabel `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `idx_users_number` (`number`),
  ADD UNIQUE KEY `idx_users_reff_code` (`reff_code`),
  ADD KEY `idx_users_reff_by` (`reff_by`);

--
-- Indeks untuk tabel `user_spins`
--
ALTER TABLE `user_spins`
  ADD PRIMARY KEY (`id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_won_at` (`won_at`);

--
-- Indeks untuk tabel `user_tasks`
--
ALTER TABLE `user_tasks`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_user_task` (`user_id`,`task_id`),
  ADD KEY `task_id` (`task_id`);

--
-- Indeks untuk tabel `withdrawals`
--
ALTER TABLE `withdrawals`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `order_id` (`order_id`),
  ADD KEY `idx_user_id` (`user_id`),
  ADD KEY `idx_bank_account_id` (`bank_account_id`),
  ADD KEY `idx_order_id` (`order_id`),
  ADD KEY `idx_status` (`status`),
  ADD KEY `idx_created_at` (`created_at`);

--
-- AUTO_INCREMENT untuk tabel yang dibuang
--

--
-- AUTO_INCREMENT untuk tabel `admins`
--
ALTER TABLE `admins`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT untuk tabel `banks`
--
ALTER TABLE `banks`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=13;

--
-- AUTO_INCREMENT untuk tabel `bank_accounts`
--
ALTER TABLE `bank_accounts`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `forums`
--
ALTER TABLE `forums`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `investments`
--
ALTER TABLE `investments`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `payments`
--
ALTER TABLE `payments`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `payment_settings`
--
ALTER TABLE `payment_settings`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT untuk tabel `products`
--
ALTER TABLE `products`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT untuk tabel `settings`
--
ALTER TABLE `settings`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT untuk tabel `spin_prizes`
--
ALTER TABLE `spin_prizes`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT untuk tabel `tasks`
--
ALTER TABLE `tasks`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=15;

--
-- AUTO_INCREMENT untuk tabel `transactions`
--
ALTER TABLE `transactions`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `users`
--
ALTER TABLE `users`
  MODIFY `id` int NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;

--
-- AUTO_INCREMENT untuk tabel `user_spins`
--
ALTER TABLE `user_spins`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `user_tasks`
--
ALTER TABLE `user_tasks`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `withdrawals`
--
ALTER TABLE `withdrawals`
  MODIFY `id` int UNSIGNED NOT NULL AUTO_INCREMENT;

-- --------------------------------------------------------

--
-- Struktur untuk view `spin_prizes_with_percentage`
--
DROP TABLE IF EXISTS `spin_prizes_with_percentage`;

CREATE ALGORITHM=UNDEFINED DEFINER=`root`@`localhost` SQL SECURITY DEFINER VIEW `spin_prizes_with_percentage`  AS SELECT `spin_prizes`.`id` AS `id`, `spin_prizes`.`amount` AS `amount`, `spin_prizes`.`code` AS `code`, `spin_prizes`.`chance_weight` AS `chance_weight`, round(((`spin_prizes`.`chance_weight` * 100.0) / (select sum(`spin_prizes`.`chance_weight`) from `spin_prizes` where (`spin_prizes`.`status` = 'Active'))),2) AS `chance_percentage`, `spin_prizes`.`status` AS `status` FROM `spin_prizes` WHERE (`spin_prizes`.`status` = 'Active') ORDER BY `spin_prizes`.`amount` ASC ;

--
-- Ketidakleluasaan untuk tabel pelimpahan (Dumped Tables)
--

--
-- Ketidakleluasaan untuk tabel `bank_accounts`
--
ALTER TABLE `bank_accounts`
  ADD CONSTRAINT `fk_bank_accounts_bank` FOREIGN KEY (`bank_id`) REFERENCES `banks` (`id`) ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_bank_accounts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Ketidakleluasaan untuk tabel `forums`
--
ALTER TABLE `forums`
  ADD CONSTRAINT `forums_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

--
-- Ketidakleluasaan untuk tabel `investments`
--
ALTER TABLE `investments`
  ADD CONSTRAINT `fk_investments_product` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_investments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Ketidakleluasaan untuk tabel `transactions`
--
ALTER TABLE `transactions`
  ADD CONSTRAINT `fk_transactions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Ketidakleluasaan untuk tabel `user_spins`
--
ALTER TABLE `user_spins`
  ADD CONSTRAINT `fk_spins_prize` FOREIGN KEY (`prize_id`) REFERENCES `spin_prizes` (`id`) ON DELETE RESTRICT,
  ADD CONSTRAINT `fk_user_spins_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Ketidakleluasaan untuk tabel `user_tasks`
--
ALTER TABLE `user_tasks`
  ADD CONSTRAINT `user_tasks_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  ADD CONSTRAINT `user_tasks_ibfk_2` FOREIGN KEY (`task_id`) REFERENCES `tasks` (`id`);

--
-- Ketidakleluasaan untuk tabel `withdrawals`
--
ALTER TABLE `withdrawals`
  ADD CONSTRAINT `fk_bank_account_id` FOREIGN KEY (`bank_account_id`) REFERENCES `bank_accounts` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT,
  ADD CONSTRAINT `fk_withdrawals_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
