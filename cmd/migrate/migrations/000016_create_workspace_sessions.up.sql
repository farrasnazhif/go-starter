CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title varchar(255) NOT NULL,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_updated
ON sessions(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS prds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  parent_prd_id UUID REFERENCES prds(id) ON DELETE SET NULL,
  version int NOT NULL,
  content_markdown text NOT NULL,
  mode varchar(40) NOT NULL DEFAULT 'standard',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  is_current boolean NOT NULL DEFAULT true,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  CONSTRAINT prds_mode_check CHECK (
    mode IN ('standard', 'developer')
  ),
  UNIQUE(session_id, version)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_prds_one_current_per_session
ON prds(session_id)
WHERE is_current = true;

CREATE INDEX IF NOT EXISTS idx_prds_session_version
ON prds(session_id, version DESC);

CREATE TABLE IF NOT EXISTS messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  type varchar(20) NOT NULL,
  content text NOT NULL,
  prd_id UUID NOT NULL REFERENCES prds(id) ON DELETE CASCADE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  CONSTRAINT messages_type_check CHECK (
    type IN ('initial', 'refine')
  )
);

CREATE INDEX IF NOT EXISTS idx_messages_session_created
ON messages(session_id, created_at ASC);
