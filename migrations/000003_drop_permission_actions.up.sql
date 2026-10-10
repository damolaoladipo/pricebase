-- 000003_drop_permission_actions.up.sql
--
-- Permission actions are stored as JSON in permissions.actions (written by
-- PERMISSION_INSERT, decoded into Permission.Actions on read). The
-- permission_actions table created by an earlier revision of 000002 was a
-- mistake: drop it if it exists. Safe to run on databases that never had
-- the table.
--
-- NOTE: the runner executes every *.sql file, so never add *.down.sql files.

DROP TABLE IF EXISTS permission_actions;
