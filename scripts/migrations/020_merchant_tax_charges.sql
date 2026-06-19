-- Merchant-level VAT and service charge configuration

ALTER TABLE merchants ADD COLUMN IF NOT EXISTS vat_percent DECIMAL(5, 2) NOT NULL DEFAULT 15;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS service_charge_percent DECIMAL(5, 2);

ALTER TABLE merchants DROP CONSTRAINT IF EXISTS chk_merchants_vat_percent;
ALTER TABLE merchants ADD CONSTRAINT chk_merchants_vat_percent CHECK (
    vat_percent >= 0 AND vat_percent <= 100
);

ALTER TABLE merchants DROP CONSTRAINT IF EXISTS chk_merchants_service_charge_percent;
ALTER TABLE merchants ADD CONSTRAINT chk_merchants_service_charge_percent CHECK (
    service_charge_percent IS NULL
    OR (service_charge_percent >= 0 AND service_charge_percent <= 100)
);
