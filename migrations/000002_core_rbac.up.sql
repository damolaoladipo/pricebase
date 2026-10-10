-- 000002_core_rbac.up.sql
--
-- Brings the schema in line with what the code actually queries
-- (internal/utils/queries.util.go, used via internal/repository and
-- internal/services) and what internal/models defines.
--
-- 000001 created only a skeletal users table (id, email, password,
-- created_at, updated_at) with no id default, so USER_INSERT (which omits
-- id and RETURNs it) fails. This migration:
--
--   1. Completes the users table from models.User db tags and gives id a
--      gen_random_uuid() default.
--   2. Creates roles (name UNIQUE is required by ROLE_INSERT's
--      ON CONFLICT (name) DO NOTHING).
--   3. Creates permissions with an actions JSONB column (written by
--      PermissionRepository.CreatePermission, read back decoded into
--      Permission.Actions). Actions are NOT a separate table.
--   4. Creates the role_permissions and user_roles join tables. Their
--      composite primary keys are required by the ON CONFLICT
--      (role_id, permission_id) / (user_id, role_id) DO NOTHING clauses.
--
-- Tables for Device/Settings/Notification/RoleTeam are intentionally NOT
-- created here: no query, repository, or service references them yet, so
-- their schema would be guesswork. Add them when code needs them.

-- ---------------------------------------------------------------------------
-- 1. Complete the users table (see models.User).
-- ---------------------------------------------------------------------------

ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS middle_name TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS logo TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_code TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS country_phone TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS alt_phone TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_type TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS saved_password TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS passcode TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS user_type TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS organization_name TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS activation_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS activation_token_expire TIMESTAMPTZ;

ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_password_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_password_token_expire TIMESTAMPTZ;

ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_code TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_code_expire TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS otp_type TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_token_expire TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS invite_status TEXT;

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_organization BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_user BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_team BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_activated BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_locked BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_phone_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_welcomed BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users ADD COLUMN IF NOT EXISTS login_limit INTEGER NOT NULL DEFAULT 0;

-- Referenced tables (countries, organizations, team roles, settings) do not
-- exist yet, so these stay as plain UUID columns with no foreign keys.
ALTER TABLE users ADD COLUMN IF NOT EXISTS country_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS organization_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS team_role_id UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS settings_id UUID;

ALTER TABLE users ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 0;

-- USER_INSERT omits id and RETURNs it, so it needs a default.
ALTER TABLE users ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- ---------------------------------------------------------------------------
-- 2. Roles (see models.Role; ROLE_INSERT / ROLE_FIND_*).
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    slug TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

-- ---------------------------------------------------------------------------
-- 3. Permissions (see models.Permission;
--    PERMISSION_INSERT / PERMISSION_FIND_*).
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    "group" TEXT NOT NULL UNIQUE,
    actions JSONB NOT NULL DEFAULT '[]',
    slug TEXT,
    updated_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 0
);

-- ---------------------------------------------------------------------------
-- 4. Join tables (see models.RolePermission; ROLE_PERMISSION_INSERT,
--    USER_ROLE_INSERT, PERMISSIONS_BY_ROLE_ID).
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);
