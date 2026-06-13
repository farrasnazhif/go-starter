CREATE TABLE IF NOT EXISTS otp_codes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email citext NOT NULL,
  code varchar(6) NOT NULL,
  purpose varchar(50) NOT NULL, -- 'registration' or 'login'
  expires_at timestamp(0) with time zone NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_codes_email ON otp_codes(email);
CREATE INDEX idx_otp_codes_expires_at ON otp_codes(expires_at);