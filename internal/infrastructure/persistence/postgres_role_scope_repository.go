package persistence

import (
	"context"
	"database/sql"
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/role/entity"
	"sipon-api/internal/domain/role/valueobject"
)

type PostgresRoleScopeRepository struct {
	db *sql.DB
}

func NewPostgresRoleScopeRepository(db *sql.DB) *PostgresRoleScopeRepository {
	return &PostgresRoleScopeRepository{db: db}
}

func (r *PostgresRoleScopeRepository) Save(ctx context.Context, scope *entity.RoleScope) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO role_scopes (id, role_id, scope_type, scope_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		scope.ID, scope.RoleID, string(scope.ScopeType), scope.ScopeValue, scope.CreatedAt, scope.UpdatedAt,
	)
	if err != nil {
		return domainerr.New("DOMAIN_ROLE_SCOPE_PERSISTENCE_FAILED")
	}
	return nil
}

func (r *PostgresRoleScopeRepository) Delete(ctx context.Context, id string) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `DELETE FROM role_scopes WHERE id = $1`, id)
	if err != nil {
		return domainerr.New("DOMAIN_ROLE_SCOPE_PERSISTENCE_FAILED")
	}
	return nil
}

func (r *PostgresRoleScopeRepository) FindByID(ctx context.Context, id string) (*entity.RoleScope, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, role_id, scope_type, scope_value, created_at, updated_at
		FROM role_scopes WHERE id = $1`, id)

	var sid, roleID, scopeType, scopeValue string
	var createdAt, updatedAt time.Time
	if err := row.Scan(&sid, &roleID, &scopeType, &scopeValue, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New("DOMAIN_ROLE_SCOPE_NOT_FOUND")
		}
		return nil, domainerr.New("DOMAIN_ROLE_SCOPE_QUERY_FAILED")
	}
	return &entity.RoleScope{
		ID: sid, RoleID: roleID, ScopeType: valueobject.ScopeType(scopeType),
		ScopeValue: scopeValue, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func (r *PostgresRoleScopeRepository) FindByRoleID(ctx context.Context, roleID string) ([]*entity.RoleScope, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, role_id, scope_type, scope_value, created_at, updated_at
		FROM role_scopes WHERE role_id = $1 ORDER BY created_at`, roleID)
	if err != nil {
		return nil, domainerr.New("DOMAIN_ROLE_SCOPE_QUERY_FAILED")
	}
	defer rows.Close()

	var scopes []*entity.RoleScope
	for rows.Next() {
		var sid, rid, scopeType, scopeValue string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&sid, &rid, &scopeType, &scopeValue, &createdAt, &updatedAt); err != nil {
			return nil, domainerr.New("DOMAIN_ROLE_SCOPE_QUERY_FAILED")
		}
		scopes = append(scopes, &entity.RoleScope{
			ID: sid, RoleID: rid, ScopeType: valueobject.ScopeType(scopeType),
			ScopeValue: scopeValue, CreatedAt: createdAt, UpdatedAt: updatedAt,
		})
	}
	return scopes, rows.Err()
}
