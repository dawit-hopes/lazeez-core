-- Optional menu discount (percentage or fixed amount)

ALTER TABLE menus ADD COLUMN IF NOT EXISTS discount_type VARCHAR(20);
ALTER TABLE menus ADD COLUMN IF NOT EXISTS discount_value DECIMAL(10, 2);

ALTER TABLE menus DROP CONSTRAINT IF EXISTS chk_menus_discount;
ALTER TABLE menus ADD CONSTRAINT chk_menus_discount CHECK (
    (discount_type IS NULL AND discount_value IS NULL)
    OR (
        discount_type IN ('percentage', 'fixed')
        AND discount_value IS NOT NULL
        AND discount_value > 0
    )
);
