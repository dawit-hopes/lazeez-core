-- Master menu: nullable branch_id on menus + per-branch availability overrides

ALTER TABLE menus ALTER COLUMN branch_id DROP NOT NULL;

CREATE TABLE IF NOT EXISTS branch_menu_overrides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id UUID NOT NULL,
    menu_id UUID NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
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
CREATE INDEX IF NOT EXISTS idx_menus_master ON menus(created_at) WHERE is_deleted = FALSE AND branch_id IS NULL;

COMMENT ON TABLE branch_menu_overrides IS 'Per-branch availability overrides for master menu items';
COMMENT ON COLUMN menus.branch_id IS 'NULL for master menu items; set for branch-only items';
