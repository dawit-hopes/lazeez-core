-- Bookings: guest check-in/check-out records for physical rooms (single_rooms).

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    guest_name VARCHAR(255) NOT NULL,
    guest_phone VARCHAR(50) NOT NULL,
    number_of_nights INTEGER NOT NULL DEFAULT 1,
    check_in_date TIMESTAMP WITH TIME ZONE NOT NULL,
    check_out_date TIMESTAMP WITH TIME ZONE NOT NULL,
    passcode VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_bookings_room FOREIGN KEY (room_id) REFERENCES single_rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

-- A physical room can only have one active booking at a time.
CREATE UNIQUE INDEX IF NOT EXISTS uq_bookings_active_room
    ON bookings (room_id)
    WHERE status = 'active' AND is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_bookings_room_id
    ON bookings(room_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_bookings_branch_id
    ON bookings(branch_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_bookings_status
    ON bookings(status)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE bookings IS 'Guest check-in/check-out records for physical rooms (single_rooms)';
COMMENT ON COLUMN bookings.status IS 'Booking lifecycle: active (checked in), checked_out, or cancelled';
COMMENT ON COLUMN bookings.passcode IS 'Server-generated guest passcode issued at check-in';

CREATE TRIGGER update_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
