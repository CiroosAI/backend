-- Migration: create_settings_table.sql
CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    logo TEXT NOT NULL,
    min_withdraw DECIMAL(15,2) NOT NULL,
    max_withdraw DECIMAL(15,2) NOT NULL,
    withdraw_charge DECIMAL(15,2) NOT NULL,
    maintenance BOOLEAN NOT NULL DEFAULT FALSE,
    closed_register BOOLEAN NOT NULL DEFAULT FALSE,
    link_cs TEXT NOT NULL,
    link_group TEXT NOT NULL,
    link_app TEXT NOT NULL
);