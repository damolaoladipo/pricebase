-- 000005_user_role_snapshot.up.sql
--
-- Stores each user's role snapshot on the row: role_ids mirrors the
-- user_roles links, permissions mirrors the derived UserPermission set.
-- Both are written only by attachSuperRole (recomputed every boot via
-- ExtractPermissions), and read straight off the row — no joins at read
-- time. Runtime RBAC editors must recompute these columns.

ALTER TABLE users ADD COLUMN IF NOT EXISTS role_ids JSONB NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions JSONB NOT NULL DEFAULT '[]';
