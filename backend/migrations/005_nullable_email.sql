-- Clerk users sign in via OAuth; we don't store their email locally.
-- Drop the NOT NULL constraint so GetOrCreateByClerkID can insert with clerk_user_id only.
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
