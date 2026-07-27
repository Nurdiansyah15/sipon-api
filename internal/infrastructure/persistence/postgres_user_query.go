package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
)

// ── Struct & Constructor ──────────────────────────────────────────────────────

type PostgresUserQuery struct{ db *sql.DB }

func NewPostgresUserQuery(db *sql.DB) *PostgresUserQuery {
	return &PostgresUserQuery{db: db}
}

// ── Query Methods ─────────────────────────────────────────────────────────────

// ListUsers mengembalikan daftar user (flat) untuk admin listing dengan filter
// optional status/role_id/search dan pagination. Tidak menyertakan roles untuk
// menghindari N+1 — role summary di-load hanya pada get-by-id.
func (q *PostgresUserQuery) ListUsers(ctx context.Context, query port.UserListReadQuery) ([]port.UserReadItem, dto.Meta, error) {
	clauses, args := []string{"u.deleted_at IS NULL"}, make([]any, 0, 4)
	if s := strings.TrimSpace(query.Status); s != "" {
		args = append(args, strings.ToUpper(s))
		clauses = append(clauses, fmt.Sprintf("u.status = $%d", len(args)))
	}
	if rid := strings.TrimSpace(query.RoleID); rid != "" {
		args = append(args, rid)
		clauses = append(clauses, fmt.Sprintf("EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role_id = $%d AND ur.is_active = TRUE AND (ur.expired_at IS NULL OR ur.expired_at > NOW()))", len(args)))
	}
	if s := strings.TrimSpace(query.Search); s != "" {
		like := "%" + s + "%"
		args = append(args, like)
		clauses = append(clauses, fmt.Sprintf("(u.username ILIKE $%d OR u.email ILIKE $%d OR COALESCE(u.fullname, '') ILIKE $%d)", len(args), len(args), len(args)))
	}
	whereSQL := strings.Join(clauses, " AND ")

	var total int64
	if err := execFromContext(ctx, q.db).QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM users u WHERE %s`, whereSQL), args...).Scan(&total); err != nil {
		return nil, dto.Meta{}, fmt.Errorf("count users: %w", err)
	}

	limit, offset, currentPage, sortColumn, sortType := resolvePaginationParams(
		query.PaginationParams,
		50,
		100,
		map[string]string{
			"username":    "u.username",
			"email":       "u.email",
			"status":      "u.status",
			"created_at":  "u.created_at",
			"updated_at":  "u.updated_at",
			"last_login_at": "u.last_login_at",
		},
		"u.created_at",
		"DESC",
	)
	rows, err := execFromContext(ctx, q.db).QueryContext(ctx, fmt.Sprintf(`SELECT u.id, u.username, u.fullname, u.email, u.phone, u.status, u.created_at, u.updated_at, u.last_login_at FROM users u WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d`, whereSQL, sortColumn, sortType, len(args)+1, len(args)+2), append(args, limit, offset)...)
	if err != nil {
		return nil, dto.Meta{}, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	items := make([]port.UserReadItem, 0, limit)
	for rows.Next() {
		item, err := scanUserReadItem(rows)
		if err != nil {
			return nil, dto.Meta{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, dto.Meta{}, fmt.Errorf("iterate users: %w", err)
	}
	totalPages := int64(0)
	if total > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}
	return items, dto.Meta{CurrentPage: currentPage, PerPage: int64(limit), Total: total, TotalPages: totalPages}, nil
}

// ListActiveRoleSummariesByUserID mengembalikan ringkasan assignment role aktif
// user. Dipakai oleh get_user untuk mengisi field roles tanpa N+1 list.
func (q *PostgresUserQuery) ListActiveRoleSummariesByUserID(ctx context.Context, userID string) ([]port.UserRoleSummaryReadItem, error) {
	rows, err := execFromContext(ctx, q.db).QueryContext(ctx, `
		SELECT ur.id, ur.role_id, r.name, ur.scope_type, ur.scope_id, ur.is_active
		FROM user_roles ur
		INNER JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.is_active = TRUE AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
		ORDER BY ur.assigned_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user role summaries: %w", err)
	}
	defer rows.Close()

	items := make([]port.UserRoleSummaryReadItem, 0)
	for rows.Next() {
		var scopeID sql.NullString
		var item port.UserRoleSummaryReadItem
		if err := rows.Scan(&item.ID, &item.RoleID, &item.RoleName, &item.ScopeType, &scopeID, &item.IsActive); err != nil {
			return nil, fmt.Errorf("scan user role summary: %w", err)
		}
		if scopeID.Valid {
			item.ScopeID = &scopeID.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user role summaries: %w", err)
	}
	return items, nil
}

// ── Scan Helpers ──────────────────────────────────────────────────────────────

func scanUserReadItem(scanner interface{ Scan(dest ...any) error }) (port.UserReadItem, error) {
	var fullname, phone sql.NullString
	var lastLoginAt sql.NullTime
	var item port.UserReadItem
	if err := scanner.Scan(&item.ID, &item.Username, &fullname, &item.Email, &phone, &item.Status, &item.CreatedAt, &item.UpdatedAt, &lastLoginAt); err != nil {
		return port.UserReadItem{}, err
	}
	if fullname.Valid {
		item.Fullname = &fullname.String
	}
	if phone.Valid && phone.String != "" {
		item.Phone = &phone.String
	}
	if lastLoginAt.Valid {
		t := lastLoginAt.Time
		item.LastLoginAt = &t
	}
	return item, nil
}