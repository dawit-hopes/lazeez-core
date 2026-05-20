-- Attach orphan master menu rows (branch_id IS NULL, merchant_id IS NULL) to the sole merchant.
-- Safe for single-restaurant installs; multi-tenant DBs need manual backfill per merchant.

UPDATE menus
SET merchant_id = (
    SELECT id FROM merchants WHERE is_deleted = FALSE ORDER BY created_at LIMIT 1
)
WHERE branch_id IS NULL
  AND merchant_id IS NULL
  AND is_deleted = FALSE
  AND (SELECT COUNT(*) FROM merchants WHERE is_deleted = FALSE) = 1;
