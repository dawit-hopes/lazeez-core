ALTER TABLE branch_menu_overrides
    ADD COLUMN IF NOT EXISTS is_excluded BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_branch_menu_overrides_excluded
    ON branch_menu_overrides(branch_id, menu_id) WHERE is_deleted = FALSE AND is_excluded = TRUE;
