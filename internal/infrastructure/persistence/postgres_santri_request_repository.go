package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"sipon-api/internal/app/dto"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/entity"
	"sipon-api/internal/domain/santri/repository"
)

type PostgresSantriRequestRepository struct {
	db *sql.DB
}

func NewPostgresSantriRequestRepository(db *sql.DB) *PostgresSantriRequestRepository {
	return &PostgresSantriRequestRepository{db: db}
}

func (r *PostgresSantriRequestRepository) Save(ctx context.Context, req *entity.SantriRequest) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO santri_requests (id, user_id, nis, status, notes, reviewed_by, reviewed_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		req.ID, req.UserID, req.NIS, string(req.Status), req.Notes, req.ReviewedBy, req.ReviewedAt, req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		return domainerr.Wrap(constant.CodeSantriRequestPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriRequestRepository) Update(ctx context.Context, req *entity.SantriRequest) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE santri_requests SET
			nis=$2, status=$3, notes=$4, reviewed_by=$5, reviewed_at=$6, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`,
		req.ID, req.NIS, string(req.Status), req.Notes, req.ReviewedBy, req.ReviewedAt,
	)
	if err != nil {
		return domainerr.Wrap(constant.CodeSantriRequestPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriRequestRepository) FindByID(ctx context.Context, id string) (*entity.SantriRequest, error) {
	return r.findOne(ctx, `SELECT * FROM santri_requests WHERE id=$1 AND deleted_at IS NULL`, id)
}

func (r *PostgresSantriRequestRepository) FindPendingByUserID(ctx context.Context, userID string) (*entity.SantriRequest, error) {
	return r.findOne(ctx, `SELECT * FROM santri_requests WHERE user_id=$1 AND status='pending' AND deleted_at IS NULL`, userID)
}

func (r *PostgresSantriRequestRepository) FindByStatus(ctx context.Context, status string) ([]*entity.SantriRequest, error) {
	return r.findMany(ctx, `SELECT * FROM santri_requests WHERE status=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, status)
}

var santriRequestSortable = map[string]string{
	"created_at": "created_at",
	"status":     "status",
}

func (r *PostgresSantriRequestRepository) List(ctx context.Context, query repository.SantriRequestListQuery) (*repository.SantriRequestListResult, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}

	if query.Status != nil && *query.Status != "" {
		args = append(args, *query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM santri_requests " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriRequestQueryFailed, err)
	}

	limit, offset, currentPage, sortColumn, sortType := resolvePaginationParams(dto.PaginationParams{
		Page:     query.Page,
		Limit:    query.Limit,
		SortBy:   query.SortBy,
		SortType: query.SortType,
	}, 10, 100, santriRequestSortable, "created_at", "DESC")

	orderClause := " ORDER BY " + sortColumn + " " + sortType

	args = append(args, limit, offset)
	selectQuery := fmt.Sprintf("SELECT * FROM santri_requests %s%s LIMIT $%d OFFSET $%d", where, orderClause, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriRequestQueryFailed, err)
	}
	defer rows.Close()

	var items []*entity.SantriRequest
	for rows.Next() {
		sr, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, sr)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriRequestQueryFailed, err)
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &repository.SantriRequestListResult{
		Items: items,
		Meta: repository.SantriListMeta{
			CurrentPage: currentPage,
			PerPage:     int64(limit),
			Total:       total,
			TotalPages:  totalPages,
		},
	}, nil
}

func (r *PostgresSantriRequestRepository) findOne(ctx context.Context, query string, args ...any) (*entity.SantriRequest, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanRow(row)
}

func (r *PostgresSantriRequestRepository) findMany(ctx context.Context, query string, args ...any) ([]*entity.SantriRequest, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriRequestQueryFailed, err)
	}
	defer rows.Close()

	var items []*entity.SantriRequest
	for rows.Next() {
		sr, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, sr)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriRequestQueryFailed, err)
	}
	return items, nil
}

func (r *PostgresSantriRequestRepository) scanRow(scanner interface{ Scan(...any) error }) (*entity.SantriRequest, error) {
	var sr entity.SantriRequest
	var status string
	err := scanner.Scan(
		&sr.ID, &sr.UserID, &sr.NIS, &status, &sr.Notes,
		&sr.ReviewedBy, &sr.ReviewedAt,
		&sr.CreatedAt, &sr.UpdatedAt, &sr.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New(constant.CodeSantriRequestNotFound)
		}
		return nil, domainerr.Wrap(constant.CodeSantriRequestQueryFailed, err)
	}
	sr.Status = constant.SantriRequestStatus(status)
	return &sr, nil
}
