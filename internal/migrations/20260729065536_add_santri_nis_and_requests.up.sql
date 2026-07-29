-- 20260729065536_add_santri_nis_and_requests.up.sql
-- Domain: Kesantrian
-- Adds NIS column to santri, updates user_identities constraint for NIS kind,
-- and creates santri_requests table.

-- ── santri: add NIS column ────────────────────────────────────────────────────
ALTER TABLE santri ADD COLUMN IF NOT EXISTS nis VARCHAR(10) UNIQUE;

-- ── user_identities: update CHECK constraint to include NIS kind ──────────────
ALTER TABLE user_identities DROP CONSTRAINT IF EXISTS user_identities_kind_check;
ALTER TABLE user_identities ADD CONSTRAINT user_identities_kind_check CHECK (kind IN ('EMAIL', 'PHONE', 'USERNAME', 'NIS'));

-- ── santri_requests ───────────────────────────────────────────────────────────
CREATE TABLE santri_requests (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nis         VARCHAR(10),
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    notes       TEXT,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_santri_requests_user_pending ON santri_requests (user_id) WHERE status = 'pending' AND deleted_at IS NULL;
CREATE INDEX idx_santri_requests_status ON santri_requests (status);
