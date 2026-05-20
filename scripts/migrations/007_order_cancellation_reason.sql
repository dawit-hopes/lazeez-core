ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;

COMMENT ON COLUMN orders.cancellation_reason IS 'Reason provided when order was declined/cancelled by branch staff';
