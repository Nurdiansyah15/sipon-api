-- 20260726120000_create_auth_rbac_tables.down.sql

DROP TABLE IF EXISTS verification_codes;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS user_identities;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at_timestamp();

DROP EXTENSION IF EXISTS "pg_trgm";
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
