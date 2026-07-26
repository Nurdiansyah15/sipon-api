package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	roleconstant "sipon-api/internal/domain/role/constant"
)

// ── Struct & Constructor ──────────────────────────────────────────────────────

type PostgresRoleQuery struct{ db *sql.DB }

func NewPostgresRoleQuery(db *sql.DB) *PostgresRoleQuery {
	return &PostgresRoleQuery{db: db}
}

// ── Query Methods ─────────────────────────────────────────────────────────────

func (q *PostgresRoleQuery) ListRoles(ctx context.Context, query port.RoleListReadQuery) ([]port.RoleReadItem, dto.Meta, error) {
	clauses, args := []string{"1=1"}, make([]any, 0, 4)
	if query.RoleType != "" {
		args = append(args, query.RoleType)
		clauses = append(clauses, fmt.Sprintf("r.role_type = $%d", len(args)))
	}
	if query.ScopeType != "" {
		args = append(args, query.ScopeType)
		clauses = append(clauses, fmt.Sprintf("r.scope_type = $%d", len(args)))
	}
	if query.Assignable != nil {
		args = append(args, *query.Assignable)
		clauses = append(clauses, fmt.Sprintf("r.assignable = $%d", len(args)))
	}
	whereSQL := strings.Join(clauses, " AND ")
	var total int64
	if err := execFromContext(ctx, q.db).QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM roles r WHERE %s`, whereSQL), args...).Scan(&total); err != nil {
		return nil, dto.Meta{}, fmt.Errorf("count roles: %w", err)
	}
	limit, offset, currentPage, sortColumn, sortType := resolvePaginationParams(
		query.PaginationParams,
		50,
		100,
		map[string]string{
			"name":         "r.name",
			"display_name": "r.display_name",
			"role_type":    "r.role_type",
			"scope_type":   "r.scope_type",
			"assignable":   "r.assignable",
			"created_at":   "r.created_at",
			"updated_at":   "r.updated_at",
		},
		"r.name",
		"ASC",
	)
	rows, err := execFromContext(ctx, q.db).QueryContext(ctx, fmt.Sprintf(`SELECT r.id, r.name, r.display_name, r.description, r.role_type, r.scope_type, r.assignable, r.created_at, r.updated_at FROM roles r WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`, whereSQL, sortColumn, sortType, len(args)+1, len(args)+2), append(args, limit, offset)...)
	if err != nil {
		return nil, dto.Meta{}, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()
	items := make([]port.RoleReadItem, 0, limit)
	for rows.Next() {
		item, err := scanRoleReadItem(rows)
		if err != nil {
			return nil, dto.Meta{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, dto.Meta{}, fmt.Errorf("iterate roles: %w", err)
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}
	return items, dto.Meta{CurrentPage: currentPage, PerPage: int64(limit), Total: total, TotalPages: totalPages}, nil
}

func (q *PostgresRoleQuery) ListUserRoles(ctx context.Context, query port.UserRoleListReadQuery) ([]port.UserRoleReadItem, dto.Meta, error) {
	clauses, args := []string{"1=1"}, make([]any, 0, 6)
	if query.UserID != "" {
		args = append(args, query.UserID)
		clauses = append(clauses, fmt.Sprintf("ur.user_id = $%d", len(args)))
	}
	if query.RoleID != "" {
		args = append(args, query.RoleID)
		clauses = append(clauses, fmt.Sprintf("ur.role_id = $%d", len(args)))
	}
	if query.ScopeType != "" {
		args = append(args, query.ScopeType)
		clauses = append(clauses, fmt.Sprintf("ur.scope_type = $%d", len(args)))
	}
	if query.ScopeID != "" {
		args = append(args, query.ScopeID)
		clauses = append(clauses, fmt.Sprintf("ur.scope_id = $%d", len(args)))
	}
	if query.IsActive != nil {
		args = append(args, *query.IsActive)
		clauses = append(clauses, fmt.Sprintf("ur.is_active = $%d", len(args)))
	}
	whereSQL := strings.Join(clauses, " AND ")
	var total int64
	if err := execFromContext(ctx, q.db).QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM user_roles ur WHERE %s`, whereSQL), args...).Scan(&total); err != nil {
		return nil, dto.Meta{}, fmt.Errorf("count user roles: %w", err)
	}
	limit, offset, currentPage, sortColumn, sortType := resolvePaginationParams(
		query.PaginationParams,
		50,
		100,
		map[string]string{
			"assigned_at": "ur.assigned_at",
			"username":    "u.username",
			"email":       "u.email",
			"role_name":   "r.name",
			"scope_type":  "ur.scope_type",
		},
		"ur.assigned_at",
		"DESC",
	)
	rows, err := execFromContext(ctx, q.db).QueryContext(ctx, fmt.Sprintf(`SELECT ur.id, ur.user_id, u.fullname, u.username, u.email, u.phone, ur.role_id, r.id, r.name, r.display_name, r.role_type, r.assignable, ur.scope_type, ur.scope_id, ur.assigned_at, ur.assigned_by, ur.expired_at, ur.is_active, ur.deactivated_at FROM user_roles ur INNER JOIN users u ON u.id=ur.user_id INNER JOIN roles r ON r.id=ur.role_id WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`, whereSQL, sortColumn, sortType, len(args)+1, len(args)+2), append(args, limit, offset)...)
	if err != nil {
		return nil, dto.Meta{}, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()
	items := make([]port.UserRoleReadItem, 0, limit)
	for rows.Next() {
		item, err := scanUserRoleReadItem(rows)
		if err != nil {
			return nil, dto.Meta{}, err
		}
		item.Permissions = permissionKeyStrings(roleconstant.PermissionsForRole(roleconstant.RoleName(item.Role.Name)))
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, dto.Meta{}, fmt.Errorf("iterate user roles: %w", err)
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}
	return items, dto.Meta{CurrentPage: currentPage, PerPage: int64(limit), Total: total, TotalPages: totalPages}, nil
}

func permissionKeyStrings(keys []roleconstant.PermissionKey) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, string(k))
	}
	return out
}

// ── Read Item Scan Helpers ────────────────────────────────────────────────────

func scanRoleReadItem(scanner interface{ Scan(dest ...any) error }) (port.RoleReadItem, error) {
	var description sql.NullString
	var item port.RoleReadItem
	if err := scanner.Scan(&item.ID, &item.Name, &item.DisplayName, &description, &item.RoleType, &item.ScopeType, &item.Assignable, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return port.RoleReadItem{}, err
	}
	if description.Valid {
		item.Description = &description.String
	}
	return item, nil
}

func scanUserRoleReadItem(scanner interface{ Scan(dest ...any) error }) (port.UserRoleReadItem, error) {
	var fullname, email, phone, scopeID, assignedBy sql.NullString
	var username string
	var expiredAt, deactivatedAt sql.NullTime
	var item port.UserRoleReadItem
	if err := scanner.Scan(&item.ID, &item.UserID, &fullname, &username, &email, &phone, &item.RoleID, &item.Role.ID, &item.Role.Name, &item.Role.DisplayName, &item.Role.RoleType, &item.Role.Assignable, &item.ScopeType, &scopeID, &item.AssignedAt, &assignedBy, &expiredAt, &item.IsActive, &deactivatedAt); err != nil {
		return port.UserRoleReadItem{}, err
	}
	item.User.ID = item.UserID
	if fullname.Valid {
		item.User.Name = &fullname.String
	} else {
		item.User.Name = &username
	}
	if email.Valid {
		item.User.Email = &email.String
	}
	if phone.Valid {
		item.User.Phone = &phone.String
	}
	if scopeID.Valid {
		item.ScopeID = &scopeID.String
	}
	if assignedBy.Valid {
		item.AssignedBy = &assignedBy.String
	}
	if expiredAt.Valid {
		item.ExpiredAt = &expiredAt.Time
	}
	if deactivatedAt.Valid {
		item.DeactivatedAt = &deactivatedAt.Time
	}
	return item, nil
}
