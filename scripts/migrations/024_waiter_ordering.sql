-- Waiter table ordering: a parallel, non-payment, waiter-driven dine-in flow.
-- Adds staff roles + waiter PINs, station tagging on the menu, station/fulfillment
-- state on order items, waiter linkage on orders, the per-table "check" (tab), and a
-- per-merchant bill print policy. The guest QR + Chapa flow is left untouched.

-- ============================================
-- USERS: new staff roles + waiter PIN hash
-- ============================================
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN (
    'super_admin',
    'branch_manager',
    'super_branch_admin',
    'front_desk_agent',
    'room_service_staff',
    'waiter',
    'kitchen_staff',
    'barista',
    'cashier'
));

ALTER TABLE users ADD COLUMN IF NOT EXISTS passcode_hash VARCHAR(255);
COMMENT ON COLUMN users.passcode_hash IS 'Hashed PIN for individual waiters (PIN-only records, no password)';

-- ============================================
-- CATEGORIES: preparation station
-- ============================================
ALTER TABLE categories ADD COLUMN IF NOT EXISTS station VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE categories DROP CONSTRAINT IF EXISTS chk_categories_station;
ALTER TABLE categories ADD CONSTRAINT chk_categories_station CHECK (station IN ('', 'kitchen', 'bar'));
COMMENT ON COLUMN categories.station IS 'Default preparation station for items in this category (kitchen/bar), empty if unset';

-- ============================================
-- MENUS: optional per-item station override (NULL = inherit from category)
-- ============================================
ALTER TABLE menus ADD COLUMN IF NOT EXISTS station VARCHAR(20);
ALTER TABLE menus DROP CONSTRAINT IF EXISTS chk_menus_station;
ALTER TABLE menus ADD CONSTRAINT chk_menus_station CHECK (station IS NULL OR station IN ('kitchen', 'bar'));
COMMENT ON COLUMN menus.station IS 'Per-item station override; NULL inherits the category station';

-- ============================================
-- ORDERS: source + waiter/check linkage
-- ============================================
ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_source VARCHAR(20) NOT NULL DEFAULT 'client';
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_order_source;
ALTER TABLE orders ADD CONSTRAINT chk_orders_order_source CHECK (order_source IN ('client', 'waiter'));

ALTER TABLE orders ADD COLUMN IF NOT EXISTS waiter_id UUID;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS check_id UUID;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_waiter;
ALTER TABLE orders ADD CONSTRAINT fk_orders_waiter FOREIGN KEY (waiter_id) REFERENCES users(id) ON DELETE SET NULL;

COMMENT ON COLUMN orders.order_source IS 'client (guest QR + Chapa) or waiter (tablet, PIN-attributed, no payment)';
COMMENT ON COLUMN orders.waiter_id IS 'PIN-resolved waiter who placed a waiter order';
COMMENT ON COLUMN orders.check_id IS 'Table check (tab) this waiter order belongs to';

CREATE INDEX IF NOT EXISTS idx_orders_check_id ON orders(check_id) WHERE is_deleted = FALSE AND check_id IS NOT NULL;

-- ============================================
-- ORDER ITEMS: station routing + fulfillment state
-- ============================================
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS station VARCHAR(20) NOT NULL DEFAULT 'kitchen';
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS chk_order_items_station;
ALTER TABLE order_items ADD CONSTRAINT chk_order_items_station CHECK (station IN ('kitchen', 'bar'));

ALTER TABLE order_items ADD COLUMN IF NOT EXISTS item_status VARCHAR(20) NOT NULL DEFAULT 'sent';
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS chk_order_items_item_status;
ALTER TABLE order_items ADD CONSTRAINT chk_order_items_item_status CHECK (item_status IN ('sent', 'accepted', 'preparing', 'ready', 'cancelled'));

COMMENT ON COLUMN order_items.station IS 'Preparation station this line routes to (kitchen/bar)';
COMMENT ON COLUMN order_items.item_status IS 'Fulfillment state: sent, accepted, preparing, ready, cancelled';

CREATE INDEX IF NOT EXISTS idx_order_items_station_status ON order_items(station, item_status) WHERE is_deleted = FALSE;

-- ============================================
-- MERCHANTS: bill print policy (readiness gate strictness)
-- ============================================
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS bill_print_policy VARCHAR(20) NOT NULL DEFAULT 'strict';
ALTER TABLE merchants DROP CONSTRAINT IF EXISTS chk_merchants_bill_print_policy;
ALTER TABLE merchants ADD CONSTRAINT chk_merchants_bill_print_policy CHECK (bill_print_policy IN ('strict', 'lenient'));
COMMENT ON COLUMN merchants.bill_print_policy IS 'strict = bill only when all items ready; lenient = when no item still sent';

-- ============================================
-- TABLE CHECKS: per-seating tab (folio analog) settled by the cashier (cash only)
-- ============================================
CREATE TABLE IF NOT EXISTS table_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id UUID NOT NULL,
    table_number INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    vat_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    service_charge_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(50) NOT NULL DEFAULT '',
    settled_by UUID,
    settled_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_table_checks_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_table_checks_settled_by FOREIGN KEY (settled_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_table_checks_status CHECK (status IN ('open', 'closed', 'void'))
);

-- At most one open check per table within a branch.
CREATE UNIQUE INDEX IF NOT EXISTS uq_table_checks_open_table
    ON table_checks (branch_id, table_number)
    WHERE status = 'open' AND is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_table_checks_branch_id ON table_checks(branch_id) WHERE is_deleted = FALSE;

COMMENT ON TABLE table_checks IS 'Per-seating tab aggregating waiter orders for a table; cashier settles it cash-only';
COMMENT ON COLUMN table_checks.status IS 'Check lifecycle: open, closed (settled), or void';

-- orders.check_id references table_checks (added after the table exists).
ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_check;
ALTER TABLE orders ADD CONSTRAINT fk_orders_check FOREIGN KEY (check_id) REFERENCES table_checks(id) ON DELETE SET NULL;

DROP TRIGGER IF EXISTS update_table_checks_updated_at ON table_checks;
CREATE TRIGGER update_table_checks_updated_at
    BEFORE UPDATE ON table_checks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
