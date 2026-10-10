-- 000006_string_ids.up.sql
--
-- IDs are application-generated 24-char hex strings
-- (utils.GenerateID), not database UUIDs. This migration converts every id
-- and reference column from UUID to TEXT, drops the gen_random_uuid()
-- defaults (inserts now supply the id explicitly), and re-creates the
-- foreign keys. Existing UUID values convert losslessly via USING ::text.
--
-- NOTE: one-way door. New 24-char ids cannot cast back to UUID.

-- Drop foreign keys before altering column types.
ALTER TABLE role_permissions DROP CONSTRAINT IF EXISTS role_permissions_role_id_fkey;
ALTER TABLE role_permissions DROP CONSTRAINT IF EXISTS role_permissions_permission_id_fkey;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_user_id_fkey;
ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS user_roles_role_id_fkey;

-- users
ALTER TABLE users ALTER COLUMN id DROP DEFAULT;
ALTER TABLE users ALTER COLUMN id TYPE TEXT USING id::text;
ALTER TABLE users ALTER COLUMN country_id TYPE TEXT USING country_id::text;
ALTER TABLE users ALTER COLUMN organization_id TYPE TEXT USING organization_id::text;
ALTER TABLE users ALTER COLUMN team_role_id TYPE TEXT USING team_role_id::text;
ALTER TABLE users ALTER COLUMN settings_id TYPE TEXT USING settings_id::text;

-- roles
ALTER TABLE roles ALTER COLUMN id DROP DEFAULT;
ALTER TABLE roles ALTER COLUMN id TYPE TEXT USING id::text;

-- permissions
ALTER TABLE permissions ALTER COLUMN id DROP DEFAULT;
ALTER TABLE permissions ALTER COLUMN id TYPE TEXT USING id::text;

-- role_permissions
ALTER TABLE role_permissions ALTER COLUMN role_id TYPE TEXT USING role_id::text;
ALTER TABLE role_permissions ALTER COLUMN permission_id TYPE TEXT USING permission_id::text;

-- user_roles
ALTER TABLE user_roles ALTER COLUMN user_id TYPE TEXT USING user_id::text;
ALTER TABLE user_roles ALTER COLUMN role_id TYPE TEXT USING role_id::text;

-- Re-create foreign keys on the TEXT columns.
ALTER TABLE role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE;
ALTER TABLE role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE;
ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_role_id_fkey
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE;
