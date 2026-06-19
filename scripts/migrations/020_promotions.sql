-- Merchant-scoped promotions for guest-facing catalog

CREATE TABLE IF NOT EXISTS promotions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    banner_image TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_promotions_merchant FOREIGN KEY (merchant_id) REFERENCES merchants(id) ON DELETE CASCADE,
    CONSTRAINT chk_promotions_date_range CHECK (end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_promotions_merchant_start
    ON promotions(merchant_id, start_date DESC)
    WHERE is_deleted = FALSE;
