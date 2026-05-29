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
    branch_type VARCHAR(50) NOT NULL DEFAULT 'restaurant',
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
    merchant_id UUID,
    is_locked BOOLEAN DEFAULT FALSE,
    is_first_login BOOLEAN DEFAULT TRUE,
    logging_attempts INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_users_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE SET NULL,
    CONSTRAINT fk_users_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE SET NULL,
    CONSTRAINT chk_users_role CHECK (role IN (
        'super_admin',
        'branch_manager',
        'super_branch_admin',
        'front_desk_agent',
        'room_service_staff'
    ))
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
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    ingredients TEXT[],
    category_id UUID NOT NULL,
    branch_id UUID,
    merchant_id UUID,
    preparation_time DECIMAL(10,2) NOT NULL DEFAULT 0,
    is_fasting BOOLEAN DEFAULT FALSE,
    is_available BOOLEAN DEFAULT TRUE,
    is_deleted BOOLEAN DEFAULT FALSE,
    modifiers TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_menus_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_menus_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE,
    CONSTRAINT chk_menus_ingredients CHECK (array_length(ingredients, 1) > 0)
);

-- Create index on name for faster lookups
CREATE INDEX IF NOT EXISTS idx_menus_name ON menus(name) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_menus_created_at ON menus(created_at) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_menus_master ON menus(created_at) WHERE is_deleted = FALSE AND branch_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_menus_merchant_id ON menus(merchant_id) WHERE is_deleted = FALSE AND branch_id IS NULL;

-- ============================================
-- BRANCH MENU OVERRIDES (master menu per-branch availability)
-- ============================================
CREATE TABLE IF NOT EXISTS branch_menu_overrides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id UUID NOT NULL,
    menu_id UUID NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    is_excluded BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_branch_menu_overrides_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_branch_menu_overrides_menu FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
    CONSTRAINT uq_branch_menu_overrides_branch_menu UNIQUE (branch_id, menu_id)
);

CREATE INDEX IF NOT EXISTS idx_branch_menu_overrides_branch_id ON branch_menu_overrides(branch_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_branch_menu_overrides_menu_id ON branch_menu_overrides(menu_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_branch_menu_overrides_excluded ON branch_menu_overrides(branch_id, menu_id) WHERE is_deleted = FALSE AND is_excluded = TRUE;

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
    icon TEXT,
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
COMMENT ON TABLE menus IS 'Stores menu information; branch_id NULL = master menu item';
COMMENT ON TABLE branch_menu_overrides IS 'Per-branch availability overrides for master menu items';
COMMENT ON TABLE categories IS 'Stores category information with name and icon';
COMMENT ON TABLE ingredients IS 'Stores ingredient names and icon';

COMMENT ON COLUMN users.role IS 'User role: super_admin, branch_manager, super_branch_admin, front_desk_agent, or room_service_staff';
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


-- ============================================
-- MODIFIER OPTIONS TABLE (options for modifier groups; IDs stored in modifier_groups.options)
-- ============================================
CREATE TABLE IF NOT EXISTS modifier_options (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    price_adjustment DECIMAL(10, 2) NOT NULL DEFAULT 0,
    is_default BOOLEAN DEFAULT FALSE,
    is_available BOOLEAN DEFAULT TRUE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_modifier_options_name ON modifier_options(name) WHERE is_deleted = FALSE;

-- ============================================
-- MODIFIER GROUPS TABLE (option IDs stored in options TEXT[]; group IDs stored in menus.modifiers)
-- ============================================
CREATE TABLE IF NOT EXISTS modifier_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    selection_type VARCHAR(50) NOT NULL,
    is_required BOOLEAN DEFAULT FALSE,
    min_selections INTEGER DEFAULT 0,
    max_selections INTEGER DEFAULT 0,
    options TEXT[] DEFAULT '{}',
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_modifier_groups_name ON modifier_groups(name) WHERE is_deleted = FALSE;

COMMENT ON TABLE modifier_options IS 'Selectable options for modifier groups (e.g. size, extras)';
COMMENT ON TABLE modifier_groups IS 'Modifier groups for menus (e.g. Size, Extras); option IDs in options array';

-- ============================================
-- ORDERS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_number INTEGER NOT NULL,
    table_number INTEGER NOT NULL,
    branch_id UUID NOT NULL,
    session_key VARCHAR(255),
    order_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    cancellation_reason TEXT,
    total DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(50),
    payment_status VARCHAR(50) DEFAULT 'pending',
    payment_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    payment_amount DECIMAL(10, 2) DEFAULT 0,
    payment_currency VARCHAR(10) DEFAULT 'ETB',
    payment_transaction_id VARCHAR(255),
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_orders_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_branch_order_number ON orders(branch_id, order_number) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_orders_branch_id ON orders(branch_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_orders_session_key ON orders(session_key) WHERE is_deleted = FALSE AND session_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_branch_status_date ON orders(branch_id, order_status, created_at DESC) WHERE is_deleted = FALSE;

-- ============================================
-- ORDER ITEMS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL,
    menu_item_id UUID NOT NULL,
    modifier_options TEXT[] DEFAULT '{}',
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_order_items_menu FOREIGN KEY (menu_item_id) REFERENCES menus(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE orders IS 'Stores customer orders; session_key identifies client (QR code), branch_id for restaurant';
COMMENT ON TABLE order_items IS 'Line items for each order; links to menus (menu_item_id)';
COMMENT ON COLUMN orders.order_number IS 'Short 6-digit display number shown to guests and staff (100000-999999)';
COMMENT ON COLUMN orders.session_key IS 'Guest session key from client_sessions; used for client get/list';
COMMENT ON COLUMN orders.order_status IS 'pending, processing, ready, completed, cancelled';

CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_order_items_updated_at
    BEFORE UPDATE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();


-- ============================================
-- TABLES TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS tables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(100) NOT NULL,
    branch_id UUID NOT NULL,
    reference VARCHAR(255) NOT NULL,
    qr_code TEXT NOT NULL,
    qr_version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    active_order_id UUID,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_tables_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tables_branch_id ON tables(branch_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tables_created_at ON tables(created_at) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tables_reference ON tables(reference) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_tables_status ON tables(status) WHERE is_deleted = FALSE;

-- ============================================
-- CLIENT SESSIONS TABLE
-- ============================================
CREATE TABLE IF NOT EXISTS client_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_key VARCHAR(255) NOT NULL,
    table_reference VARCHAR(255) NOT NULL,
    table_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_client_sessions_table FOREIGN KEY (table_id) REFERENCES tables(id) ON DELETE CASCADE,
    CONSTRAINT fk_client_sessions_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_client_sessions_session_key
    ON client_sessions(session_key)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_client_sessions_table_reference
    ON client_sessions(table_reference)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_client_sessions_expires_at
    ON client_sessions(expires_at)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE client_sessions IS 'Ephemeral guest sessions created when a table QR is scanned';
COMMENT ON COLUMN client_sessions.session_key IS 'Random client identifier used for order APIs (distinct from table QR reference)';
COMMENT ON COLUMN client_sessions.table_reference IS 'Permanent table QR reference used to load the menu';
COMMENT ON COLUMN client_sessions.expires_at IS 'Session expiry; new scan after expiry creates a new session';

CREATE TRIGGER update_client_sessions_updated_at
    BEFORE UPDATE ON client_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();