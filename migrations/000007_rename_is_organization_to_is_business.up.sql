-- 000007_rename_is_organization_to_is_business.up.sql
--
-- The model (models.User) and all queries reference is_business, but
-- 000002 created the column as is_organization. Rename it to match.

ALTER TABLE users RENAME COLUMN is_organization TO is_business;
