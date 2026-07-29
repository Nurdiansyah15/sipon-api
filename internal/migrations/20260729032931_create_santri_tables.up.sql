-- 20260729032931_create_santri_tables.up.sql
-- Domain: Kesantrian
-- Tables: santri, santri_dokumen

-- ── santri ────────────────────────────────────────────────────────────────────
CREATE TABLE santri (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Data Pribadi
    nickname          VARCHAR(100),
    program           VARCHAR(100),
    "option"          VARCHAR(2)   CHECK ("option" IN ('1', '2')),
    hobby             VARCHAR(200),
    purpose           VARCHAR(200),
    motivation_entry  VARCHAR(500),
    pob               VARCHAR(200),
    dob               DATE,
    blood             VARCHAR(5),

    -- Kontak
    address       TEXT,
    sub_district  VARCHAR(200),
    district      VARCHAR(200),
    province      VARCHAR(200),
    postal_code   VARCHAR(10),

    -- Pondok Sebelumnya
    previous_pondok_name    VARCHAR(200),
    previous_pondok_address VARCHAR(200),
    previous_pondok_div     VARCHAR(200),
    previous_pondok_time    VARCHAR(100),

    -- Kependudukan
    nik     VARCHAR(20),
    no_kk   VARCHAR(20),
    nisn    VARCHAR(10),
    no_kip  VARCHAR(20),
    no_kks  VARCHAR(20),
    no_pkh  VARCHAR(20),

    -- Pendidikan
    workplace  VARCHAR(200),
    department VARCHAR(200),

    -- Data Keluarga: Status Rumah
    home_status VARCHAR(100),

    -- Data Keluarga: Ayah
    father           VARCHAR(200),
    father_pn        VARCHAR(20),
    father_nik       VARCHAR(20),
    father_job       VARCHAR(200),
    father_graduate  VARCHAR(200),
    father_income    VARCHAR(50),

    -- Data Keluarga: Ibu
    mother           VARCHAR(200),
    mother_pn        VARCHAR(20),
    mother_nik       VARCHAR(20),
    mother_job       VARCHAR(200),
    mother_graduate  VARCHAR(200),
    mother_income    VARCHAR(50),

    -- Data Keluarga: Wali
    guardian_relationship VARCHAR(200),
    guardian              VARCHAR(200),
    guardian_pn           VARCHAR(20),
    guardian_nik          VARCHAR(20),
    guardian_job          VARCHAR(200),
    guardian_graduate     VARCHAR(200),
    guardian_income       VARCHAR(50),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_santri_user_id ON santri (user_id);
CREATE INDEX idx_santri_nik    ON santri (nik);
CREATE INDEX idx_santri_nisn   ON santri (nisn);

-- ── santri_dokumen ────────────────────────────────────────────────────────────
CREATE TABLE santri_dokumen (
    id                UUID PRIMARY KEY,
    santri_id         UUID NOT NULL REFERENCES santri(id) ON DELETE CASCADE,
    kind              VARCHAR(30) NOT NULL CHECK (kind IN ('surat_pernyataan', 'ktp', 'kk', 'mutasi', 'pembayaran')),
    key               TEXT NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    original_filename VARCHAR(500),
    mime_type         VARCHAR(200),
    size              BIGINT,
    notes             TEXT,
    verified_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_santri_dokumen_santri_id ON santri_dokumen (santri_id);
CREATE INDEX idx_santri_dokumen_kind     ON santri_dokumen (kind);
CREATE INDEX idx_santri_dokumen_status   ON santri_dokumen (status);
