package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"sipon-api/internal/infrastructure/persistence"
)

type dbExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func execOrTx(ctx context.Context, db *sql.DB) dbExec {
	if tx := persistence.TxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// MustInsertRole inserts a minimal role row and returns its ID.
// Satisfies FK constraints in tests for repositories that reference the roles table.
func MustInsertRole(ctx context.Context, t *testing.T, db *sql.DB) string {
	t.Helper()
	id := uuid.New().String()
	name := fmt.Sprintf("role_%s", id[:8])
	_, err := execOrTx(ctx, db).ExecContext(ctx, `
		INSERT INTO roles (id, name, display_name, role_type, scope_type, assignable)
		VALUES ($1, $2, $3, 'system', 'global', true)`,
		id, name, name,
	)
	require.NoError(t, err)
	return id
}

// MustInsertUser inserts a minimal user row (no credentials) and returns its ID.
// Use only to satisfy FK constraints for tests outside the user repository.
// To test user persistence itself, use PostgresUserRepository.Save directly.
func MustInsertUser(ctx context.Context, t *testing.T, db *sql.DB) string {
	t.Helper()
	id := uuid.New().String()
	short := id[:8]
	_, err := execOrTx(ctx, db).ExecContext(ctx, `
		INSERT INTO users (id, username, email, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'ACTIVE', NOW(), NOW())`,
		id, "usr_"+short, "usr_"+short+"@example.com",
	)
	require.NoError(t, err)
	return id
}
