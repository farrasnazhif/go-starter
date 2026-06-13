ALTER TABLE otp_codes
  ADD COLUMN IF NOT EXISTS verified_at timestamp(0) with time zone;

CREATE INDEX IF NOT EXISTS idx_otp_codes_verified
ON otp_codes(email, purpose, verified_at);
