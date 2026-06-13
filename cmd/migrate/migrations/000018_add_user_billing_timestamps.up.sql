ALTER TABLE users
  ADD COLUMN IF NOT EXISTS subscription_created_at timestamp(0) with time zone,
  ADD COLUMN IF NOT EXISTS expired_at timestamp(0) with time zone;
