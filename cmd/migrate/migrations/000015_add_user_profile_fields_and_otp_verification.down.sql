DROP INDEX IF EXISTS idx_otp_codes_verified;

ALTER TABLE otp_codes
  DROP COLUMN IF EXISTS verified_at;
