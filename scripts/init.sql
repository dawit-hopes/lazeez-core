-- Lazeez Core Database Schema
-- This script creates all necessary tables for the application

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- MERCHANTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    logo TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on name and created_at for faster lookups and sorting
CREATE INDEX IF NOT EXISTS idx_merchants_name ON merchants(name) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_merchants_created_at ON merchants(created_at) WHERE is_deleted = FALSE;

-- ============================================
-- BRANCHES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID NOT NULL,
    branch_name VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_branches_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_branches_merchant_id ON branches(merchant_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_branches_phone_number ON branches(phone_number) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_branches_created_at ON branches(created_at) WHERE is_deleted = FALSE;

-- ============================================
-- USERS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    full_name VARCHAR(100) NOT NULL,
    password VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'branch_manager',
    branch_id UUID,
    is_locked BOOLEAN DEFAULT FALSE,
    is_first_login BOOLEAN DEFAULT TRUE,
    logging_attempts INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_users_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE SET NULL,
    CONSTRAINT chk_users_role CHECK (role IN ('super_admin', 'branch_manager'))
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_users_branch_id ON users(branch_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at) WHERE is_deleted = FALSE;

-- create a super admin user
INSERT INTO users (phone_number, full_name, role, password, is_locked, is_first_login, logging_attempts) VALUES ('251945557307', 'Super Admin', 'super_admin', '$2a$10$8qgRwxx8tZC.t2DVs0jy6u1w4Au4sLz61V5ZVPxdU4V6vFEsiQseC', FALSE, FALSE, 0);

-- ============================================
-- MENUS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS menus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    image TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on name for faster lookups
CREATE INDEX IF NOT EXISTS idx_menus_name ON menus(name) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_menus_created_at ON menus(created_at) WHERE is_deleted = FALSE;

-- ============================================
-- CATEGORIES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    icon TEXT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on name for faster lookups
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_categories_created_at ON categories(created_at) WHERE is_deleted = FALSE;

-- ============================================
-- INGREDIENTS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ingredients_name ON ingredients(name) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_ingredients_created_at ON ingredients(created_at) WHERE is_deleted = FALSE;

-- ============================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================
-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for each table
CREATE TRIGGER update_merchants_updated_at
    BEFORE UPDATE ON merchants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_menus_updated_at
    BEFORE UPDATE ON menus
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ingredients_updated_at
    BEFORE UPDATE ON ingredients
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================
COMMENT ON TABLE merchants IS 'Stores merchant/restaurant information';
COMMENT ON TABLE branches IS 'Stores branch locations for merchants';
COMMENT ON TABLE users IS 'Stores user accounts with authentication information';
COMMENT ON TABLE menus IS 'Stores menu information';
COMMENT ON TABLE categories IS 'Stores category information with name and icon';
COMMENT ON TABLE ingredients IS 'Stores ingredient names';

COMMENT ON COLUMN users.role IS 'User role: super admin or branch_manager';
COMMENT ON COLUMN users.is_locked IS 'Indicates if user account is locked';
COMMENT ON COLUMN users.is_first_login IS 'Indicates if this is the user''s first login';
COMMENT ON COLUMN users.logging_attempts IS 'Number of failed login attempts';


-- ============================================
-- SESSIONS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    refresh_token TEXT NOT NULL,
    access_token TEXT NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE    
);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at) WHERE is_deleted = FALSE;



