-- Harden booking passcodes: store a bcrypt hash instead of plaintext and add
-- brute-force attempt tracking used by guest room-order authentication.

ALTER TABLE bookings ADD COLUMN IF NOT EXISTS passcode_hash VARCHAR(255);
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS passcode_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS passcode_locked_until TIMESTAMP WITH TIME ZONE;

-- Existing plaintext passcodes cannot be rehashed; drop the legacy column.
-- (Active guests would need a re-issued passcode.)
ALTER TABLE bookings DROP COLUMN IF EXISTS passcode;

COMMENT ON COLUMN bookings.passcode_hash IS 'bcrypt hash of the guest passcode issued at check-in (plaintext returned only once)';
COMMENT ON COLUMN bookings.passcode_attempts IS 'Consecutive failed passcode attempts for room-order authentication';
COMMENT ON COLUMN bookings.passcode_locked_until IS 'Timestamp until which passcode verification is locked after too many failures';
