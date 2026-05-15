-- Scope master menu items per restaurant (merchant)

ALTER TABLE menus ADD COLUMN IF NOT EXISTS merchant_id UUID;

ALTER TABLE menus
    ADD CONSTRAINT fk_menus_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_menus_merchant_id ON menus(merchant_id) WHERE is_deleted = FALSE AND branch_id IS NULL;

-- Backfill merchant_id on branch-owned items from their branch
UPDATE menus m
SET merchant_id = b.merchant_id
FROM branches b
WHERE m.branch_id = b.id AND m.merchant_id IS NULL;
