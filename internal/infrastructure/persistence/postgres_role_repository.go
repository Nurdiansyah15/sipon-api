package persistence

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	domainerr "sipon-api/internal/domain/errors"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"

	"github.com/jackc/pgx/v5/pgconn"
)

// ── Structs & Constructors ────────────────────────────────────────────────────

type PostgresRoleRepository struct{ db *sql.DB }
type PostgresUserRoleRepository struct{ db *sql.DB }

func NewPostgresRoleRepository(db *sql.DB) *PostgresRoleRepository {
	return &PostgresRoleRepository{db: db}
}
func NewPostgresUserRoleRepository(db *sql.DB) *PostgresUserRoleRepository {
	return &PostgresUserRoleRepository{db: db}
}

// ── Role Repository ───────────────────────────────────────────────────────────

func (r *PostgresRoleRepository) Save(ctx context.Context, role *roleentity.Role) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO roles (id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		role.ID, string(role.Name), role.DisplayName, role.Description, string(role.RoleType), string(role.ScopeType), role.Assignable, role.CreatedAt, role.UpdatedAt,
	)
	if err != nil {
		return mapRolePersistenceError(err)
	}
	return nil
}

func (r *PostgresRoleRepository) Update(ctx context.Context, role *roleentity.Role) error {
	result, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		UPDATE roles SET name=$1, display_name=$2, description=$3, role_type=$4, scope_type=$5, assignable=$6, updated_at=$7 WHERE id=$8`,
		string(role.Name), role.DisplayName, role.Description, string(role.RoleType), string(role.ScopeType), role.Assignable, role.UpdatedAt, role.ID,
	)
	if err != nil {
		return mapRolePersistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return domainerr.New(roleconstant.CodeRoleNotFound)
	}
	return nil
}

func (r *PostgresRoleRepository) Delete(ctx context.Context, id string) error {
	result, err := execFromContext(ctx, r.db).ExecContext(ctx, `DELETE FROM roles WHERE id=$1`, id)
	if err != nil {
		return mapRolePersistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return domainerr.New(roleconstant.CodeRoleNotFound)
	}
	return nil
}

func (r *PostgresRoleRepository) FindByID(ctx context.Context, id string) (*roleentity.Role, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx, `SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles WHERE id=$1`, id)
	return scanRoleEntity(row)
}

func (r *PostgresRoleRepository) FindByName(ctx context.Context, name roleconstant.RoleName) (*roleentity.Role, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx, `SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles WHERE LOWER(name)=LOWER($1)`, string(name))
	return scanRoleEntity(row)
}

func (r *PostgresRoleRepository) ListByType(ctx context.Context, roleType roleconstant.RoleType) ([]*roleentity.Role, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `SELECT id, name, display_name, description, role_type, scope_type, assignable, created_at, updated_at FROM roles WHERE role_type=$1 ORDER BY name ASC`, string(roleType))
	if err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeRoleQueryFailed, err)
	}
	defer rows.Close()
	items := make([]*roleentity.Role, 0)
	for rows.Next() {
		item, err := scanRoleEntityRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeRoleQueryFailed, err)
	}
	return items, nil
}

// ── UserRole Repository ───────────────────────────────────────────────────────

func (r *PostgresUserRoleRepository) Save(ctx context.Context, userRole *roleentity.UserRole) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, deactivated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userRole.ID, userRole.UserID, userRole.RoleID, string(userRole.ScopeType), userRole.ScopeID, userRole.AssignedAt, nullableString(userRole.AssignedBy), userRole.ExpiredAt, userRole.IsActive, userRole.DeactivatedAt,
	)
	if err != nil {
		return mapUserRolePersistenceError(err)
	}
	return nil
}

func (r *PostgresUserRoleRepository) Update(ctx context.Context, userRole *roleentity.UserRole) error {
	result, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		UPDATE user_roles SET role_id=$1, scope_type=$2, scope_id=$3, expired_at=$4, is_active=$5, deactivated_at=$6 WHERE id=$7`,
		userRole.RoleID, string(userRole.ScopeType), userRole.ScopeID, userRole.ExpiredAt, userRole.IsActive, userRole.DeactivatedAt, userRole.ID,
	)
	if err != nil {
		return mapUserRolePersistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return domainerr.New(roleconstant.CodeUserRoleNotFound)
	}
	return nil
}

func (r *PostgresUserRoleRepository) Delete(ctx context.Context, id string) error {
	result, err := execFromContext(ctx, r.db).ExecContext(ctx, `DELETE FROM user_roles WHERE id=$1`, id)
	if err != nil {
		return mapUserRolePersistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return domainerr.New(roleconstant.CodeUserRoleNotFound)
	}
	return nil
}

func (r *PostgresUserRoleRepository) FindByID(ctx context.Context, id string) (*roleentity.UserRole, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx, `SELECT id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, deactivated_at FROM user_roles WHERE id=$1`, id)
	return scanUserRoleEntity(row)
}

func (r *PostgresUserRoleRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*roleentity.UserRole, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `SELECT id, user_id, role_id, scope_type, scope_id, assigned_at, assigned_by, expired_at, is_active, deactivated_at FROM user_roles WHERE user_id=$1 AND is_active=TRUE AND (expired_at IS NULL OR expired_at > NOW()) ORDER BY assigned_at DESC`, userID)
	if err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
	}
	defer rows.Close()
	items := make([]*roleentity.UserRole, 0)
	for rows.Next() {
		item, err := scanUserRoleEntityRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
	}
	return items, nil
}

func (r *PostgresUserRoleRepository) ListActiveUserIDsByRoleName(ctx context.Context, roleName roleconstant.RoleName) ([]string, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT DISTINCT ur.user_id
		FROM user_roles ur
		JOIN roles rl ON rl.id = ur.role_id
		WHERE LOWER(rl.name) = LOWER($1) AND ur.is_active = TRUE AND (ur.expired_at IS NULL OR ur.expired_at > NOW())`,
		string(roleName),
	)
	if err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
	}
	return ids, nil
}

// ── Entity Scan Helpers ───────────────────────────────────────────────────────

func scanRoleEntity(row *sql.Row) (*roleentity.Role, error) {
	var description sql.NullString
	var item roleentity.Role
	if err := row.Scan(&item.ID, &item.Name, &item.DisplayName, &description, &item.RoleType, &item.ScopeType, &item.Assignable, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerr.New(roleconstant.CodeRoleNotFound)
		}
		return nil, domainerr.Wrap(roleconstant.CodeRoleQueryFailed, err)
	}
	if description.Valid {
		item.Description = &description.String
	}
	return &item, nil
}

func scanRoleEntityRows(rows *sql.Rows) (*roleentity.Role, error) {
	var description sql.NullString
	var item roleentity.Role
	if err := rows.Scan(&item.ID, &item.Name, &item.DisplayName, &description, &item.RoleType, &item.ScopeType, &item.Assignable, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeRoleQueryFailed, err)
	}
	if description.Valid {
		item.Description = &description.String
	}
	return &item, nil
}

func scanUserRoleEntity(row *sql.Row) (*roleentity.UserRole, error) {
	var scopeID, assignedBy sql.NullString
	var expiredAt, deactivatedAt sql.NullTime
	var item roleentity.UserRole
	if err := row.Scan(&item.ID, &item.UserID, &item.RoleID, &item.ScopeType, &scopeID, &item.AssignedAt, &assignedBy, &expiredAt, &item.IsActive, &deactivatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerr.New(roleconstant.CodeUserRoleNotFound)
		}
		return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
	}
	if scopeID.Valid {
		item.ScopeID = &scopeID.String
	}
	if assignedBy.Valid {
		item.AssignedBy = assignedBy.String
	}
	if expiredAt.Valid {
		item.ExpiredAt = &expiredAt.Time
	}
	if deactivatedAt.Valid {
		item.DeactivatedAt = &deactivatedAt.Time
	}
	return &item, nil
}

func scanUserRoleEntityRows(rows *sql.Rows) (*roleentity.UserRole, error) {
	var scopeID, assignedBy sql.NullString
	var expiredAt, deactivatedAt sql.NullTime
	var item roleentity.UserRole
	if err := rows.Scan(&item.ID, &item.UserID, &item.RoleID, &item.ScopeType, &scopeID, &item.AssignedAt, &assignedBy, &expiredAt, &item.IsActive, &deactivatedAt); err != nil {
		return nil, domainerr.Wrap(roleconstant.CodeUserRoleQueryFailed, err)
	}
	if scopeID.Valid {
		item.ScopeID = &scopeID.String
	}
	if assignedBy.Valid {
		item.AssignedBy = assignedBy.String
	}
	if expiredAt.Valid {
		item.ExpiredAt = &expiredAt.Time
	}
	if deactivatedAt.Valid {
		item.DeactivatedAt = &deactivatedAt.Time
	}
	return &item, nil
}

// ── Persistence Error Mappers ─────────────────────────────────────────────────

func mapRolePersistenceError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domainerr.Wrap(roleconstant.CodeRoleDuplicateName, err)
	}
	return domainerr.Wrap(roleconstant.CodeRolePersistenceFailed, err)
}

func mapUserRolePersistenceError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domainerr.Wrap(roleconstant.CodeUserRoleDuplicate, err)
	}
	return domainerr.Wrap(roleconstant.CodeUserRolePersistenceFailed, err)
}

// ── Shared Helpers ────────────────────────────────────────────────────────────

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
