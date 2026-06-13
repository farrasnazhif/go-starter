ALTER TABLE users
  DROP COLUMN IF EXISTS expired_at,
  DROP COLUMN IF EXISTS subscription_created_at;
