package persistence

import (
	"context"
	"database/sql"
	"errors"

	domainerr "sipon-api/internal/domain/errors"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRolePermissionRepository struct{ db *sql.DB }

func NewPostgresRolePermissionRepository(db *sql.DB) *PostgresRolePermissionRepository {
	return &PostgresRolePermissionRepository{db: db}
}

func (r *PostgresRolePermissionRepository) Save(ctx context.Context, rp *roleentity.RolePermission) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO role_permissions (id, role_id, permission_key, assigned_at, assigned_by, notes)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		rp.ID, rp.RoleID, string(rp.PermissionKey), rp.AssignedAt, nullableString(rp.AssignedBy), rp.Notes,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerr.Wrap(roleconstant.CodeRolePermissionDuplicate, err)
		}
		return domainerr.Wrap(roleconstant.CodeRolePermissionPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresRolePermissionRepository) Delete(ctx context.Context, roleID string, permissionKey roleconstant.PermissionKey) error {
	result, err := execFromContext(ctx, r.db).ExecContext(ctx,
		`DELETE FROM role_permissions WHERE role_id=$1 AND permission_key=$2`,
		roleID, string(permissionKey),
	)
	if err != nil {
		return domainerr.Wrap(roleconstant.CodeRolePermissionPersistenceFailed, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return domainerr.New(roleconstant.CodeRolePermissionNotFound)
	}
	return nil
}

func (r *PostgresRolePermissionRepository) ListByRoleID(ctx context.Context, roleID string) ([]*roleentity.RolePermission, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, role_id, permission_key, assigned_at, assigned_by, notes
		FROM role_permissions
		WHERE role_id=$1
		ORDER BY permission_key ASC`, roleID)
	if err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeRolePermissionQueryFailed, err)
	}
	defer rows.Close()

	items := make([]*roleentity.RolePermission, 0)
	for rows.Next() {
		var (
			id, rID, key string
			assignedBy   sql.NullString
			notes        sql.NullString
			item         roleentity.RolePermission
		)
		if err := rows.Scan(&id, &rID, &key, &item.AssignedAt, &assignedBy, &notes); err != nil {
			return nil, domainerr.Wrap(roleconstant.CodeRolePermissionQueryFailed, err)
		}
		item.ID = id
		item.RoleID = rID
		item.PermissionKey = roleconstant.PermissionKey(key)
		if assignedBy.Valid {
			item.AssignedBy = assignedBy.String
		}
		if notes.Valid {
			item.Notes = &notes.String
		}
		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeRolePermissionQueryFailed, err)
	}
	return items, nil
}
