package seeders

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserSeeder struct{}

type userSeed struct {
	Username       string
	Fullname       string
	Email          string
	Password       string
	UsernameStatus string
	EmailStatus    string
}

var predefinedUsers = []userSeed{
	{
		Username:       "usergod",
		Fullname:       "User God Seed",
		Email:          "usergod@sipon.dev",
		Password:       "Usergod1234",
		UsernameStatus: "VERIFIED",
		EmailStatus:    "VERIFIED",
	},
}

func (UserSeeder) Name() string {
	return "user"
}

func (UserSeeder) Run(ctx context.Context, db *sql.DB) error {
	for _, u := range predefinedUsers {
		if err := upsertUserWithRole(ctx, db, u); err != nil {
			return err
		}
	}
	return nil
}

func upsertUserWithRole(ctx context.Context, db *sql.DB, seed userSeed) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password user seeder: %w", err)
	}

	fullname := strings.TrimSpace(seed.Fullname)
	email := strings.TrimSpace(seed.Email)
	username := strings.TrimSpace(seed.Username)
	usernameStatus := strings.TrimSpace(seed.UsernameStatus)
	emailStatus := strings.TrimSpace(seed.EmailStatus)
	now := time.Now()

	var userID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO users (id, username, fullname, email, status)
		VALUES ($1, $2, $3, $4, 'ACTIVE')
		ON CONFLICT (email) DO UPDATE SET
			username = EXCLUDED.username,
			fullname = EXCLUDED.fullname,
			status = 'ACTIVE',
			deleted_at = NULL,
			updated_at = NOW()
		RETURNING id`,
		uuid.NewString(), username, fullname, email,
	).Scan(&userID); err != nil {
		return fmt.Errorf("upsert user seeder %s: %w", email, err)
	}

	credentialID, err := upsertLocalCredential(ctx, db, userID, string(hashedPassword), now)
	if err != nil {
		return err
	}

	if err := upsertIdentity(ctx, db, userID, credentialID, "EMAIL", email, emailStatus, true, now); err != nil {
		return err
	}
	if err := upsertIdentity(ctx, db, userID, credentialID, "USERNAME", username, usernameStatus, true, now); err != nil {
		return err
	}

	roleID, err := resolveUsergodRoleID(ctx, db)
	if err != nil {
		return err
	}

	if err := upsertGlobalUserRole(ctx, db, userID, roleID); err != nil {
		return err
	}

	return nil
}

func upsertLocalCredential(ctx context.Context, db *sql.DB, userID, secretHash string, now time.Time) (string, error) {
	var credentialID string
	err := db.QueryRowContext(ctx, `
		SELECT id
		FROM credentials
		WHERE user_id = $1 AND type = 'LOCAL' AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT 1`, userID,
	).Scan(&credentialID)
	if err == nil {
		if _, err := db.ExecContext(ctx, `
			UPDATE credentials
			SET secret_hash = $2,
				last_changed_at = $3,
				is_primary = TRUE,
				updated_at = NOW()
			WHERE id = $1`, credentialID, secretHash, now,
		); err != nil {
			return "", fmt.Errorf("update credential local user %s: %w", userID, err)
		}
		return credentialID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("query credential local user %s: %w", userID, err)
	}

	if err := db.QueryRowContext(ctx, `
		INSERT INTO credentials (id, user_id, type, secret_hash, last_changed_at, is_primary)
		VALUES ($1, $2, 'LOCAL', $3, $4, TRUE)
		RETURNING id`,
		uuid.NewString(), userID, secretHash, now,
	).Scan(&credentialID); err != nil {
		return "", fmt.Errorf("insert credential local user %s: %w", userID, err)
	}
	return credentialID, nil
}

func upsertIdentity(ctx context.Context, db *sql.DB, userID, credentialID, kind, value, status string, isPrimary bool, verifiedAt time.Time) error {
	if _, err := db.ExecContext(ctx, `
		INSERT INTO user_identities (id, user_id, credential_id, kind, value, status, is_primary, verified_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (kind, value) WHERE deleted_at IS NULL DO UPDATE SET
			user_id = EXCLUDED.user_id,
			credential_id = EXCLUDED.credential_id,
			status = EXCLUDED.status,
			is_primary = EXCLUDED.is_primary,
			verified_at = EXCLUDED.verified_at,
			updated_at = NOW(),
			deleted_at = NULL`,
		uuid.NewString(), userID, credentialID, kind, value, status, isPrimary, verifiedAt,
	); err != nil {
		return fmt.Errorf("upsert identity %s:%s untuk user %s: %w", kind, value, userID, err)
	}
	return nil
}

func resolveUsergodRoleID(ctx context.Context, db *sql.DB) (string, error) {
	var roleID string
	if err := db.QueryRowContext(ctx, `
		SELECT id
		FROM roles
		WHERE LOWER(name) = 'usergod'
		LIMIT 1`,
	).Scan(&roleID); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("role usergod tidak ditemukan, jalankan seeder role dulu")
		}
		return "", fmt.Errorf("query role usergod: %w", err)
	}
	return roleID, nil
}

func upsertGlobalUserRole(ctx context.Context, db *sql.DB, userID, roleID string) error {
	result, err := db.ExecContext(ctx, `
		UPDATE user_roles
		SET is_active = TRUE,
			expired_at = NULL,
			deactivated_at = NULL
		WHERE user_id = $1
		  AND role_id = $2
		  AND scope_type = 'global'
		  AND scope_id IS NULL`,
		userID, roleID,
	)
	if err != nil {
		return fmt.Errorf("update user role global user %s: %w", userID, err)
	}

	if rows, _ := result.RowsAffected(); rows > 0 {
		return nil
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, assigned_by, assigned_at, scope_type, scope_id, is_active)
		VALUES ($1, $2, $3, $2, NOW(), 'global', NULL, TRUE)`,
		uuid.NewString(), userID, roleID,
	); err != nil {
		return fmt.Errorf("insert user role global user %s: %w", userID, err)
	}

	return nil
}
