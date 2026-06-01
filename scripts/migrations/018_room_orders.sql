-- Room orders: guest food/drink orders placed from a room and charged to the
-- room bill (folio) rather than paid online. No payment columns.

CREATE TABLE IF NOT EXISTS room_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_number INTEGER NOT NULL,
    room_id UUID NOT NULL,
    booking_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    session_key VARCHAR(255) NOT NULL,
    order_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    cancellation_reason VARCHAR(255),
    total NUMERIC(10, 2) NOT NULL,
    bill_id UUID NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_room_orders_room FOREIGN KEY (room_id) REFERENCES single_rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_orders_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_orders_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_orders_bill FOREIGN KEY (bill_id) REFERENCES room_bills(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_room_orders_branch_order_number
    ON room_orders(branch_id, order_number)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_room_orders_booking_id
    ON room_orders(booking_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_room_orders_branch_id
    ON room_orders(branch_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_room_orders_session_key
    ON room_orders(session_key)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE room_orders IS 'Guest room orders charged to the room bill (folio); not paid online';
COMMENT ON COLUMN room_orders.order_status IS 'Order lifecycle: pending, processing, ready, completed, or cancelled';

CREATE TRIGGER update_room_orders_updated_at
    BEFORE UPDATE ON room_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS room_order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_order_id UUID NOT NULL,
    menu_item_id UUID NOT NULL,
    modifier_options TEXT[] DEFAULT '{}',
    quantity INTEGER NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_room_order_items_order FOREIGN KEY (room_order_id) REFERENCES room_orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_order_items_menu FOREIGN KEY (menu_item_id) REFERENCES menus(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_room_order_items_order_id
    ON room_order_items(room_order_id)
    WHERE is_deleted = FALSE;

CREATE TRIGGER update_room_order_items_updated_at
    BEFORE UPDATE ON room_order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
