package persistence

import (
	"context"
	"database/sql"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/entity"
)

type PostgresSantriDokumenRepository struct {
	db *sql.DB
}

func NewPostgresSantriDokumenRepository(db *sql.DB) *PostgresSantriDokumenRepository {
	return &PostgresSantriDokumenRepository{db: db}
}

func (r *PostgresSantriDokumenRepository) Save(ctx context.Context, dokumen *entity.SantriDokumen) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO santri_dokumen (
			id, santri_id, kind, key, status, original_filename, mime_type, size,
			notes, verified_by, verified_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		dokumen.ID, dokumen.SantriID, string(dokumen.Kind), dokumen.Key, string(dokumen.Status),
		dokumen.OriginalFilename, dokumen.MimeType, dokumen.Size,
		dokumen.Notes, dokumen.VerifiedBy, dokumen.VerifiedAt,
		dokumen.CreatedAt, dokumen.UpdatedAt,
	)
	if err != nil {
		return domainerr.Wrap(constant.CodeDokumenPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriDokumenRepository) Update(ctx context.Context, dokumen *entity.SantriDokumen) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE santri_dokumen SET
			status=$2, notes=$3, verified_by=$4, verified_at=$5, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`,
		dokumen.ID, string(dokumen.Status), dokumen.Notes, dokumen.VerifiedBy, dokumen.VerifiedAt,
	)
	if err != nil {
		return domainerr.Wrap(constant.CodeDokumenPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriDokumenRepository) FindByID(ctx context.Context, id string) (*entity.SantriDokumen, error) {
	return r.findOne(ctx, `SELECT * FROM santri_dokumen WHERE id=$1 AND deleted_at IS NULL`, id)
}

func (r *PostgresSantriDokumenRepository) FindBySantriID(ctx context.Context, santriID string) ([]*entity.SantriDokumen, error) {
	return r.findMany(ctx, `SELECT * FROM santri_dokumen WHERE santri_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, santriID)
}

func (r *PostgresSantriDokumenRepository) FindBySantriIDAndKind(ctx context.Context, santriID, kind string) ([]*entity.SantriDokumen, error) {
	return r.findMany(ctx, `SELECT * FROM santri_dokumen WHERE santri_id=$1 AND kind=$2 AND deleted_at IS NULL ORDER BY created_at DESC`, santriID, kind)
}

func (r *PostgresSantriDokumenRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE santri_dokumen SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return domainerr.Wrap(constant.CodeDokumenPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriDokumenRepository) findOne(ctx context.Context, query string, args ...any) (*entity.SantriDokumen, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	var d entity.SantriDokumen
	var kind, status string
	err := row.Scan(
		&d.ID, &d.SantriID, &kind, &d.Key, &status,
		&d.OriginalFilename, &d.MimeType, &d.Size,
		&d.Notes, &d.VerifiedBy, &d.VerifiedAt,
		&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New(constant.CodeDokumenNotFound)
		}
		return nil, domainerr.Wrap(constant.CodeDokumenQueryFailed, err)
	}
	d.Kind = constant.DokumenKind(kind)
	d.Status = constant.DokumenStatus(status)
	return &d, nil
}

func (r *PostgresSantriDokumenRepository) findMany(ctx context.Context, query string, args ...any) ([]*entity.SantriDokumen, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, domainerr.Wrap(constant.CodeDokumenQueryFailed, err)
	}
	defer rows.Close()

	var items []*entity.SantriDokumen
	for rows.Next() {
		var d entity.SantriDokumen
		var kind, status string
		if err := rows.Scan(
			&d.ID, &d.SantriID, &kind, &d.Key, &status,
			&d.OriginalFilename, &d.MimeType, &d.Size,
			&d.Notes, &d.VerifiedBy, &d.VerifiedAt,
			&d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		); err != nil {
			return nil, domainerr.Wrap(constant.CodeDokumenQueryFailed, err)
		}
		d.Kind = constant.DokumenKind(kind)
		d.Status = constant.DokumenStatus(status)
		items = append(items, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(constant.CodeDokumenQueryFailed, err)
	}
	return items, nil
}
