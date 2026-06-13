DROP INDEX IF EXISTS idx_users_paypal_subscription_id;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_credits_check;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
  DROP COLUMN IF EXISTS subscription_status,
  DROP COLUMN IF EXISTS paypal_subscription_id,
  DROP COLUMN IF EXISTS credits,
  DROP COLUMN IF EXISTS role;
