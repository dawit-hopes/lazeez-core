-- Room types: merchant master (branch_id NULL) or branch-owned rows (branch-specific or clone via parent_id)

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    merchant_id UUID NOT NULL,
    branch_id UUID,
    parent_id UUID,
    price_per_night NUMERIC(12, 2) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_rooms_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE,
    CONSTRAINT fk_rooms_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_rooms_parent FOREIGN KEY (parent_id) REFERENCES rooms(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_master_name
    ON rooms (merchant_id, name)
    WHERE branch_id IS NULL AND parent_id IS NULL AND is_deleted = FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_branch_name
    ON rooms (branch_id, name)
    WHERE branch_id IS NOT NULL AND parent_id IS NULL AND is_deleted = FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_branch_clone
    ON rooms (branch_id, parent_id)
    WHERE parent_id IS NOT NULL AND is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_rooms_merchant_master
    ON rooms (merchant_id)
    WHERE branch_id IS NULL AND parent_id IS NULL AND is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_rooms_branch_id
    ON rooms (branch_id)
    WHERE is_deleted = FALSE;

CREATE INDEX IF NOT EXISTS idx_rooms_parent_id
    ON rooms (parent_id)
    WHERE is_deleted = FALSE;

COMMENT ON TABLE rooms IS 'Hotel room types: merchant master (branch_id NULL) or branch-owned (specific type or clone of master via parent_id)';
COMMENT ON COLUMN rooms.branch_id IS 'NULL for merchant master room types; set for branch-specific or cloned types';
COMMENT ON COLUMN rooms.parent_id IS 'When set, this row is a branch clone of the master room type referenced by parent_id. Deleting a master hard-deletes its clones via ON DELETE CASCADE.';
