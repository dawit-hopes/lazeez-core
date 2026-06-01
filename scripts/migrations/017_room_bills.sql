-- Room bills (folios): one open bill per active booking that aggregates room
-- orders charged to the room and is settled by staff at/ before check-out.

CREATE TABLE IF NOT EXISTS room_bills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    total NUMERIC(10, 2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(50) NOT NULL DEFAULT '',
    settled_by UUID,
    settled_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_room_bills_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_bills_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

-- One folio per booking.
CREATE UNIQUE INDEX IF NOT EXISTS uq_room_bills_booking
    ON room_bills (booking_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_room_bills_branch_id
    ON room_bills(branch_id)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE room_bills IS 'Per-booking folio aggregating room orders charged to the room';
COMMENT ON COLUMN room_bills.status IS 'Folio lifecycle: open, settled, or void';
COMMENT ON COLUMN room_bills.total IS 'Running total of non-cancelled room orders attached to this bill';

CREATE TRIGGER update_room_bills_updated_at
    BEFORE UPDATE ON room_bills
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
