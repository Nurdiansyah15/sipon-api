-- 20260726120000_create_auth_rbac_tables.up.sql
-- Domain: Auth & Authorization (RBAC)
-- Tables: users, credentials, user_identities, roles, user_roles, verification_codes
-- Catatan: permission key & mapping role→permission TIDAK disimpan di DB — keduanya
-- didefinisikan sebagai constant di internal/domain/role/constant (lihat PermissionKey,
-- RolePermissions). Tabel `roles`/`user_roles` hanya menyimpan role & assignment-nya.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Shared trigger function: keep updated_at in sync
CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ── users ─────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id                    UUID         PRIMARY KEY,
    username              VARCHAR(30)  NOT NULL UNIQUE,
    fullname              VARCHAR(100) NULL,
    email                 VARCHAR(255) NOT NULL UNIQUE,
    phone                 VARCHAR(20)  UNIQUE,
    status                VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'BANNED')),
    failed_login_attempts INT          NOT NULL DEFAULT 0,
    locked_until          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_login_at         TIMESTAMPTZ,
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_users_username   ON users (username);
CREATE INDEX idx_users_status     ON users (status);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

-- ── credentials ───────────────────────────────────────────────────────────────
CREATE TABLE credentials (
    id              UUID        PRIMARY KEY,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            VARCHAR(30) NOT NULL CHECK (type IN ('LOCAL')),
    secret_hash     TEXT,
    last_changed_at TIMESTAMPTZ,
    is_primary      BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_credentials_user_id ON credentials (user_id);

-- ── user_identities ───────────────────────────────────────────────────────────
CREATE TABLE user_identities (
    id            UUID         PRIMARY KEY,
    user_id       UUID         NOT NULL REFERENCES users(id)       ON DELETE CASCADE,
    credential_id UUID         NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    kind          VARCHAR(20)  NOT NULL CHECK (kind IN ('EMAIL', 'PHONE', 'USERNAME')),
    value         VARCHAR(255) NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'UNVERIFIED' CHECK (status IN ('VERIFIED', 'UNVERIFIED')),
    is_primary    BOOLEAN      NOT NULL DEFAULT FALSE,
    verified_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX        idx_user_identities_user_id          ON user_identities (user_id);
CREATE INDEX        idx_user_identities_credential_id    ON user_identities (credential_id);
CREATE INDEX        idx_user_identities_kind_value       ON user_identities (kind, value);
CREATE UNIQUE INDEX uq_user_identities_kind_value_active ON user_identities (kind, value)   WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_user_identities_user_primary_kind ON user_identities (user_id, kind) WHERE is_primary = TRUE AND deleted_at IS NULL;

-- ── roles ─────────────────────────────────────────────────────────────────────
-- scope_type tetap punya 3 nilai untuk ekstensibilitas masa depan (region/community),
-- meski saat ini hanya role dengan scope 'global' yang di-seed dan dipakai.
CREATE TABLE roles (
    id           UUID         PRIMARY KEY,
    name         VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    description  TEXT,
    role_type    VARCHAR(20)  NOT NULL DEFAULT 'system' CHECK (role_type  IN ('system', 'custom')),
    scope_type   VARCHAR(20)  NOT NULL DEFAULT 'global' CHECK (scope_type IN ('global', 'region', 'community')),
    assignable   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_roles_role_type  ON roles (role_type);
CREATE INDEX idx_roles_scope_type ON roles (scope_type);

-- ── user_roles ────────────────────────────────────────────────────────────────
CREATE TABLE user_roles (
    id             UUID        PRIMARY KEY,
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id        UUID        NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_type     VARCHAR(20) NOT NULL CHECK (scope_type IN ('global', 'region', 'community')),
    scope_id       UUID,
    assigned_by    UUID        NOT NULL,
    assigned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at     TIMESTAMPTZ,
    is_active      BOOLEAN     NOT NULL DEFAULT TRUE,
    deactivated_at TIMESTAMPTZ
);

CREATE INDEX        idx_user_roles_user_id      ON user_roles (user_id);
CREATE UNIQUE INDEX uq_user_roles_active_global ON user_roles (user_id, role_id, scope_type)           WHERE is_active = TRUE AND scope_id IS NULL;
CREATE UNIQUE INDEX uq_user_roles_active_scoped ON user_roles (user_id, role_id, scope_type, scope_id) WHERE is_active = TRUE AND scope_id IS NOT NULL;

-- ── verification_codes ────────────────────────────────────────────────────────
CREATE TABLE verification_codes (
    id                 UUID        PRIMARY KEY,
    user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code               VARCHAR(6)  NOT NULL,
    purpose            VARCHAR(30) NOT NULL CHECK (purpose IN ('EMAIL_VERIFICATION', 'PHONE_VERIFICATION', 'RESET_PASSWORD', 'CHANGE_EMAIL', 'CHANGE_PHONE')),
    new_identity_value VARCHAR(255),
    expires_at         TIMESTAMPTZ NOT NULL,
    used_at            TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verif_user_purpose ON verification_codes (user_id, purpose);
