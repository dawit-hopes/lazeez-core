-- Room sessions: ephemeral guest sessions created when a room QR is scanned and
-- the booking passcode is verified. Bound to a specific active booking so a guest
-- can only order for their own room.

CREATE TABLE IF NOT EXISTS room_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_key VARCHAR(255) NOT NULL,
    room_reference VARCHAR(255) NOT NULL,
    room_id UUID NOT NULL,
    booking_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_room_sessions_room FOREIGN KEY (room_id) REFERENCES single_rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_sessions_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    CONSTRAINT fk_room_sessions_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_room_sessions_session_key
    ON room_sessions(session_key)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_room_sessions_booking_id
    ON room_sessions(booking_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_room_sessions_expires_at
    ON room_sessions(expires_at)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE room_sessions IS 'Ephemeral guest sessions created when a room QR is scanned and the booking passcode is verified';
COMMENT ON COLUMN room_sessions.session_key IS 'Random client identifier used for room-order APIs (distinct from room QR reference)';
COMMENT ON COLUMN room_sessions.booking_id IS 'Active booking the session is bound to; orders are charged to this booking';

CREATE TRIGGER update_room_sessions_updated_at
    BEFORE UPDATE ON room_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
