-- 000010_model_alignment.up.sql
--
-- Aligns the schema with internal/models (the source of truth) on a fresh
-- database. Drops the role_permissions join table: role→permission links now
-- live in roles.permissions (JSONB array of permission IDs), mirroring how
-- user→role links live in users.role_ids.

-- Role permission links move into the array column.
DROP TABLE IF EXISTS role_permissions;
ALTER TABLE roles ADD COLUMN IF NOT EXISTS permissions JSONB NOT NULL DEFAULT '[]';

-- User sub-objects and snapshots the model declares.
ALTER TABLE users ADD COLUMN IF NOT EXISTS business_name TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS business_id TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS login JSONB NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboard JSONB NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS register JSONB NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS status JSONB NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth JSONB NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS country JSONB NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS notifications_id JSONB NOT NULL DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS devices_id JSONB NOT NULL DEFAULT '[]';

-- Device registry (see models.Device).
CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT '',
    login TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    device_version TEXT NOT NULL DEFAULT '',
    device_id TEXT NOT NULL DEFAULT '',
    os JSONB NOT NULL DEFAULT '{}',
    client JSONB NOT NULL DEFAULT '{}',
    details JSONB NOT NULL DEFAULT '{}',
    source TEXT NOT NULL DEFAULT '',
    push_token TEXT NOT NULL DEFAULT '',
    is_mobile BOOLEAN NOT NULL DEFAULT FALSE,
    is_desktop BOOLEAN NOT NULL DEFAULT FALSE,
    browser TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

-- Notification records (see models.Notification).
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

-- Per-user/business settings (see models.Settings).
CREATE TABLE IF NOT EXISTS settings (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL DEFAULT '',
    login JSONB NOT NULL DEFAULT '{}',
    usage JSONB NOT NULL DEFAULT '{}',
    billing JSONB NOT NULL DEFAULT '{}',
    user_id TEXT NOT NULL DEFAULT '',
    business_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

-- Custom organization roles (see models.RoleTeam).
CREATE TABLE IF NOT EXISTS role_teams (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    slug TEXT NOT NULL DEFAULT '',
    users JSONB NOT NULL DEFAULT '[]',
    business_id TEXT NOT NULL DEFAULT '',
    permissions JSONB NOT NULL DEFAULT '[]',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);
