-- Physical rooms (individual hotel rooms of a given room type), scoped to a branch.

CREATE TABLE IF NOT EXISTS single_rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_number VARCHAR(100) NOT NULL,
    room_type_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    floor INTEGER NOT NULL DEFAULT 0,
    reference VARCHAR(255) NOT NULL,
    qr_code TEXT,
    qr_version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'vacant',
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_single_rooms_type FOREIGN KEY (room_type_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_single_rooms_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_single_rooms_branch_number
    ON single_rooms (branch_id, room_number)
    WHERE is_deleted = FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_single_rooms_reference ON single_rooms(reference);

CREATE UNIQUE INDEX IF NOT EXISTS uq_single_rooms_qr_code
    ON single_rooms(qr_code)
    WHERE qr_code IS NOT NULL AND qr_code != '';

CREATE INDEX IF NOT EXISTS idx_single_rooms_branch_id
    ON single_rooms(branch_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_single_rooms_room_type_id
    ON single_rooms(room_type_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_single_rooms_reference ON single_rooms(reference);

COMMENT ON TABLE single_rooms IS 'Individual physical hotel rooms of a given room type, scoped to a branch; each has its own QR code';
COMMENT ON COLUMN single_rooms.room_type_id IS 'Room type (rooms table) this physical room belongs to';
COMMENT ON COLUMN single_rooms.reference IS 'Permanent opaque QR reference encoded into the room QR code';

CREATE TRIGGER update_single_rooms_updated_at
    BEFORE UPDATE ON single_rooms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
