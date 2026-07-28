CREATE TABLE user_scopes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope_type  VARCHAR(50) NOT NULL CHECK (scope_type IN ('gender')),
    scope_value VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_scopes UNIQUE (user_id, scope_type, scope_value)
);
