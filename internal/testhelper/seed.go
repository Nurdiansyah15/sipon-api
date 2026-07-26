package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// MustSeedNamedRole menyisipkan role dengan nama tertentu.
// Idempotent — jika role dengan nama tersebut sudah ada, mengembalikan ID-nya.
func MustSeedNamedRole(ctx context.Context, t *testing.T, db *sql.DB, name, roleType, scopeType string) string {
	if t != nil {
		t.Helper()
	}
	var existing string
	err := db.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = $1 LIMIT 1`, name).Scan(&existing)
	if err == nil {
		return existing
	}
	id := uuid.New().String()
	displayName := name
	_, err = db.ExecContext(ctx, `
		INSERT INTO roles (id, name, display_name, role_type, scope_type, assignable)
		VALUES ($1, $2, $3, $4, $5, true)`,
		id, name, displayName, roleType, scopeType,
	)
	if t != nil {
		require.NoError(t, err, "MustSeedNamedRole: insert role %s", name)
	} else if err != nil {
		panic(fmt.Sprintf("MustSeedNamedRole %s: %v", name, err))
	}
	return id
}

// MustVerifyUserEmailIdentity menandai email identity user sebagai verified langsung via DB.
// Digunakan untuk test yang membutuhkan user dengan email sudah terverifikasi (mis. ForgotPassword).
func MustVerifyUserEmailIdentity(ctx context.Context, t *testing.T, db *sql.DB, userID string) {
	if t != nil {
		t.Helper()
	}
	_, err := db.ExecContext(ctx, `
		UPDATE user_identities
		SET status = 'VERIFIED', verified_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND kind = 'EMAIL'`,
		userID,
	)
	if t != nil {
		require.NoError(t, err, "MustVerifyUserEmailIdentity")
	} else if err != nil {
		panic(fmt.Sprintf("MustVerifyUserEmailIdentity: %v", err))
	}
}

// MustAssignUserRole menyisipkan user_role entry untuk user tertentu.
func MustAssignUserRole(ctx context.Context, t *testing.T, db *sql.DB, userID, roleID, assignerID string) {
	if t != nil {
		t.Helper()
	}
	id := uuid.New().String()
	_, err := db.ExecContext(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, scope_type, assigned_by, is_active, assigned_at)
		VALUES ($1, $2, $3, 'global', $4, true, NOW())`,
		id, userID, roleID, assignerID,
	)
	if t != nil {
		require.NoError(t, err, "MustAssignUserRole")
	} else if err != nil {
		panic(fmt.Sprintf("MustAssignUserRole: %v", err))
	}
}
