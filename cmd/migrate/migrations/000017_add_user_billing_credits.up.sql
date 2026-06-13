ALTER TABLE users
  ADD COLUMN IF NOT EXISTS role varchar(20) NOT NULL DEFAULT 'free',
  ADD COLUMN IF NOT EXISTS credits int NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS paypal_subscription_id varchar(100),
  ADD COLUMN IF NOT EXISTS subscription_status varchar(40) NOT NULL DEFAULT 'none';

ALTER TABLE users
  ADD CONSTRAINT users_role_check CHECK (role IN ('free', 'pro'));

ALTER TABLE users
  ADD CONSTRAINT users_credits_check CHECK (credits >= 0);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_paypal_subscription_id
ON users(paypal_subscription_id)
WHERE paypal_subscription_id IS NOT NULL;
