-- 20260729065536_add_santri_nis_and_requests.down.sql

DROP TABLE IF EXISTS santri_requests;

ALTER TABLE user_identities DROP CONSTRAINT IF EXISTS user_identities_kind_check;
ALTER TABLE user_identities ADD CONSTRAINT user_identities_kind_check CHECK (kind IN ('EMAIL', 'PHONE', 'USERNAME'));

ALTER TABLE santri DROP COLUMN IF EXISTS nis;
