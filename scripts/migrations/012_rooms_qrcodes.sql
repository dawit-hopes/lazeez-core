ALTER TABLE rooms ADD COLUMN IF NOT EXISTS qr_code TEXT;
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS qr_version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS reference VARCHAR(255);

UPDATE rooms
SET reference = gen_random_uuid()::text
WHERE reference IS NULL OR reference = '';

ALTER TABLE rooms ALTER COLUMN reference SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_rooms_reference ON rooms(reference);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_reference ON rooms(reference);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rooms_qr_code
    ON rooms(qr_code)
    WHERE qr_code IS NOT NULL AND qr_code != '';
