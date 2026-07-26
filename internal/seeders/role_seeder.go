package seeders

import (
	"context"
	"database/sql"
	"fmt"

	roleconstant "sipon-api/internal/domain/role/constant"

	"github.com/google/uuid"
)

type RoleSeeder struct{}

func (RoleSeeder) Name() string {
	return "role"
}

// Run meng-upsert seluruh system role dari roleconstant.DefaultRolesInit —
// satu-satunya sumber kebenaran daftar system role, supaya tidak ada daftar
// role kedua di seeder yang bisa tidak sinkron dengan constant.
func (RoleSeeder) Run(ctx context.Context, db *sql.DB) error {
	for name, meta := range roleconstant.DefaultRolesInit {
		if _, err := upsertRole(ctx, db, name, meta.DisplayName, meta.Description, meta.RoleType, meta.ScopeType, meta.Assignable); err != nil {
			return err
		}
	}
	return nil
}

func upsertRole(ctx context.Context, db *sql.DB, name roleconstant.RoleName, displayName, description string, roleType roleconstant.RoleType, scopeType roleconstant.ScopeType, assignable bool) (string, error) {
	roleID := uuid.NewString()
	var persistedID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO roles (id, name, display_name, description, role_type, scope_type, assignable)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (name) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			role_type = EXCLUDED.role_type,
			scope_type = EXCLUDED.scope_type,
			assignable = EXCLUDED.assignable,
			updated_at = NOW()
		RETURNING id`,
		roleID, name, displayName, description, roleType, scopeType, assignable,
	).Scan(&persistedID); err != nil {
		return "", fmt.Errorf("upsert role %s: %w", name, err)
	}
	return persistedID, nil
}
