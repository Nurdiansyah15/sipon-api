package persistence

import (
	"context"
	"database/sql"
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/entity"
	"sipon-api/internal/domain/user/valueobject"
)

type PostgresUserScopeRepository struct {
	db *sql.DB
}

func NewPostgresUserScopeRepository(db *sql.DB) *PostgresUserScopeRepository {
	return &PostgresUserScopeRepository{db: db}
}

func (r *PostgresUserScopeRepository) Save(ctx context.Context, scope *entity.UserScope) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO user_scopes (id, user_id, scope_type, scope_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		scope.ID, scope.UserID, string(scope.ScopeType), scope.ScopeValue, scope.CreatedAt, scope.UpdatedAt,
	)
	if err != nil {
		return domainerr.New("DOMAIN_USER_SCOPE_PERSISTENCE_FAILED")
	}
	return nil
}

func (r *PostgresUserScopeRepository) Delete(ctx context.Context, id string) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		DELETE FROM user_scopes WHERE id = $1`, id)
	if err != nil {
		return domainerr.New("DOMAIN_USER_SCOPE_PERSISTENCE_FAILED")
	}
	return nil
}

func (r *PostgresUserScopeRepository) FindByID(ctx context.Context, id string) (*entity.UserScope, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, user_id, scope_type, scope_value, created_at, updated_at
		FROM user_scopes WHERE id = $1`, id)

	var (
		sid, userID, scopeType, scopeValue string
		createdAt, updatedAt               time.Time
	)
	if err := row.Scan(&sid, &userID, &scopeType, &scopeValue, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New("DOMAIN_USER_SCOPE_NOT_FOUND")
		}
		return nil, domainerr.New("DOMAIN_USER_SCOPE_QUERY_FAILED")
	}

	return &entity.UserScope{
		ID:         sid,
		UserID:     userID,
		ScopeType:  valueobject.UserScopeType(scopeType),
		ScopeValue: scopeValue,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (r *PostgresUserScopeRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.UserScope, error) {
	rows, err := execFromContext(ctx, r.db).QueryContext(ctx, `
		SELECT id, user_id, scope_type, scope_value, created_at, updated_at
		FROM user_scopes WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, domainerr.New("DOMAIN_USER_SCOPE_QUERY_FAILED")
	}
	defer rows.Close()

	var scopes []*entity.UserScope
	for rows.Next() {
		var (
			sid, uid, scopeType, scopeValue string
			createdAt, updatedAt            time.Time
		)
		if err := rows.Scan(&sid, &uid, &scopeType, &scopeValue, &createdAt, &updatedAt); err != nil {
			return nil, domainerr.New("DOMAIN_USER_SCOPE_QUERY_FAILED")
		}
		scopes = append(scopes, &entity.UserScope{
			ID:         sid,
			UserID:     uid,
			ScopeType:  valueobject.UserScopeType(scopeType),
			ScopeValue: scopeValue,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		})
	}
	return scopes, rows.Err()
}
