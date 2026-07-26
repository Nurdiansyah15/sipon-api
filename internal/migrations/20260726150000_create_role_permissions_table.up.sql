-- 20260726150000_create_role_permissions_table.up.sql
-- Domain: Authorization (custom role permission assignment)
--
-- Permission key TIDAK punya tabel master (lihat internal/domain/role/constant —
-- daftar key valid & mapping permission role SYSTEM tetap didefinisikan di kode,
-- bukan di DB). Tabel ini HANYA menyimpan assignment permission untuk role
-- CUSTOM (role_type='custom') yang dibuat via API — makanya tidak ada FK ke
-- tabel permissions (karena memang tidak ada), permission_key divalidasi di
-- application layer terhadap constant.AllPermissionKeys().

CREATE TABLE role_permissions (
    id             UUID         PRIMARY KEY,
    role_id        UUID         NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_key VARCHAR(100) NOT NULL,
    assigned_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    assigned_by    UUID,
    notes          TEXT,
    CONSTRAINT uq_role_permissions UNIQUE (role_id, permission_key)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);
