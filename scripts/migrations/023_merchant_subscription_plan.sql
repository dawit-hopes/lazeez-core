-- Merchant subscription plan (DIGITAL_MENU or ORDERING)

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS subscription_plan VARCHAR(50) NOT NULL DEFAULT 'DIGITAL_MENU';

ALTER TABLE merchants DROP CONSTRAINT IF EXISTS chk_merchants_subscription_plan;
ALTER TABLE merchants ADD CONSTRAINT chk_merchants_subscription_plan CHECK (
    subscription_plan IN ('DIGITAL_MENU', 'ORDERING')
);
