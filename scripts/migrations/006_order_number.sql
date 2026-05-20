-- Short 6-digit display order number (per branch, unique among active orders)
ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_number INTEGER;

WITH numbered AS (
    SELECT
        id,
        100000 + (ROW_NUMBER() OVER (PARTITION BY branch_id ORDER BY created_at))::INTEGER AS num
    FROM orders
    WHERE order_number IS NULL
)
UPDATE orders o
SET order_number = numbered.num
FROM numbered
WHERE o.id = numbered.id AND o.order_number IS NULL;

UPDATE orders
SET order_number = 100000 + (ABS(hashtext(id::text)) % 900000)
WHERE order_number IS NULL;

ALTER TABLE orders ALTER COLUMN order_number SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_branch_order_number
    ON orders (branch_id, order_number)
    WHERE is_deleted = FALSE;

COMMENT ON COLUMN orders.order_number IS 'Short 6-digit display number shown to guests and staff (100000-999999)';
