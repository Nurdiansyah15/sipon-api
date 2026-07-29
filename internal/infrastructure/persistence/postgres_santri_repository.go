package persistence

import (
	"context"
	"database/sql"
	"errors"

	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"sipon-api/internal/app/dto"
	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/santri/constant"
	"sipon-api/internal/domain/santri/entity"
	"sipon-api/internal/domain/santri/repository"
	santriVO "sipon-api/internal/domain/santri/valueobject"
)

type PostgresSantriRepository struct {
	db *sql.DB
}

func NewPostgresSantriRepository(db *sql.DB) *PostgresSantriRepository {
	return &PostgresSantriRepository{db: db}
}

func (r *PostgresSantriRepository) Save(ctx context.Context, santri *entity.Santri) error {
	var nis *string
	if santri.NIS != nil {
		v := santri.NIS.Value()
		nis = &v
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO santri (
			id, user_id, nis, nickname, program, "option", hobby, purpose, motivation_entry,
			pob, dob, blood, address, sub_district, district, province, postal_code,
			previous_pondok_name, previous_pondok_address, previous_pondok_div, previous_pondok_time,
			nik, no_kk, nisn, no_kip, no_kks, no_pkh, workplace, department,
			home_status, father, father_pn, father_nik, father_job, father_graduate, father_income,
			mother, mother_pn, mother_nik, mother_job, mother_graduate, mother_income,
			guardian_relationship, guardian, guardian_pn, guardian_nik, guardian_job, guardian_graduate, guardian_income,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45,$46,$47,$48,$49,$50,$51,$52,$53,$54)`,
		santri.ID, santri.UserID, nis,
		santri.Nickname, santri.Program, santri.Option, santri.Hobby, santri.Purpose, santri.MotivationEntry,
		santri.POB, santri.DOB, santri.Blood,
		santri.Address, santri.SubDistrict, santri.District, santri.Province, santri.PostalCode,
		santri.PreviousPondokName, santri.PreviousPondokAddress, santri.PreviousPondokDiv, santri.PreviousPondokTime,
		santri.NIK, santri.NoKK, santri.NISN, santri.NoKIP, santri.NoKKS, santri.NoPKH,
		santri.Workplace, santri.Department,
		santri.HomeStatus,
		santri.Father, santri.FatherPN, santri.FatherNIK, santri.FatherJob, santri.FatherGraduate, santri.FatherIncome,
		santri.Mother, santri.MotherPN, santri.MotherNIK, santri.MotherJob, santri.MotherGraduate, santri.MotherIncome,
		santri.GuardianRelationship, santri.Guardian, santri.GuardianPN, santri.GuardianNIK, santri.GuardianJob, santri.GuardianGraduate, santri.GuardianIncome,
		santri.CreatedAt, santri.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerr.New(constant.CodeSantriDuplicate)
		}
		return domainerr.Wrap(constant.CodeSantriPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriRepository) Update(ctx context.Context, santri *entity.Santri) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE santri SET
			user_id=$2, nickname=$3, program=$4, "option"=$5, hobby=$6, purpose=$7, motivation_entry=$8,
			pob=$9, dob=$10, blood=$11, address=$12, sub_district=$13, district=$14, province=$15, postal_code=$16,
			previous_pondok_name=$17, previous_pondok_address=$18, previous_pondok_div=$19, previous_pondok_time=$20,
			nik=$21, no_kk=$22, nisn=$23, no_kip=$24, no_kks=$25, no_pkh=$26, workplace=$27, department=$28,
			home_status=$29, father=$30, father_pn=$31, father_nik=$32, father_job=$33, father_graduate=$34, father_income=$35,
			mother=$36, mother_pn=$37, mother_nik=$38, mother_job=$39, mother_graduate=$40, mother_income=$41,
			guardian_relationship=$42, guardian=$43, guardian_pn=$44, guardian_nik=$45, guardian_job=$46, guardian_graduate=$47, guardian_income=$48,
			updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`,
		santri.ID, santri.UserID,
		santri.Nickname, santri.Program, santri.Option, santri.Hobby, santri.Purpose, santri.MotivationEntry,
		santri.POB, santri.DOB, santri.Blood,
		santri.Address, santri.SubDistrict, santri.District, santri.Province, santri.PostalCode,
		santri.PreviousPondokName, santri.PreviousPondokAddress, santri.PreviousPondokDiv, santri.PreviousPondokTime,
		santri.NIK, santri.NoKK, santri.NISN, santri.NoKIP, santri.NoKKS, santri.NoPKH,
		santri.Workplace, santri.Department,
		santri.HomeStatus,
		santri.Father, santri.FatherPN, santri.FatherNIK, santri.FatherJob, santri.FatherGraduate, santri.FatherIncome,
		santri.Mother, santri.MotherPN, santri.MotherNIK, santri.MotherJob, santri.MotherGraduate, santri.MotherIncome,
		santri.GuardianRelationship, santri.Guardian, santri.GuardianPN, santri.GuardianNIK, santri.GuardianJob, santri.GuardianGraduate, santri.GuardianIncome,
	)
	if err != nil {
		return domainerr.Wrap(constant.CodeSantriPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresSantriRepository) FindByID(ctx context.Context, id string) (*entity.Santri, error) {
	return r.findOne(ctx, `SELECT * FROM santri WHERE id=$1 AND deleted_at IS NULL`, id)
}

func (r *PostgresSantriRepository) FindByUserID(ctx context.Context, userID string) (*entity.Santri, error) {
	return r.findOne(ctx, `SELECT * FROM santri WHERE user_id=$1 AND deleted_at IS NULL`, userID)
}

func (r *PostgresSantriRepository) FindByNIS(ctx context.Context, nis string) (*entity.Santri, error) {
	return r.findOne(ctx, `SELECT * FROM santri WHERE nis=$1 AND deleted_at IS NULL`, nis)
}

var santriSortableColumns = map[string]string{
	"nis":       "nis",
	"user_id":   "user_id",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func (r *PostgresSantriRepository) List(ctx context.Context, query repository.SantriListQuery) (*repository.SantriListResult, error) {
	where := "WHERE deleted_at IS NULL"
	args := []any{}

	if query.NIS != nil && *query.NIS != "" {
		args = append(args, "%"+*query.NIS+"%")
		where += fmt.Sprintf(" AND nis ILIKE $%d", len(args))
	}
	if query.UserID != nil && *query.UserID != "" {
		args = append(args, *query.UserID)
		where += fmt.Sprintf(" AND user_id = $%d", len(args))
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM santri " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriQueryFailed, err)
	}

	limit, offset, currentPage, sortColumn, sortType := resolvePaginationParams(dto.PaginationParams{
		Page:     query.Page,
		Limit:    query.Limit,
		SortBy:   query.SortBy,
		SortType: query.SortType,
	}, 10, 100, santriSortableColumns, "created_at", "DESC")

	orderClause := " ORDER BY " + sortColumn + " " + sortType

	args = append(args, limit, offset)
	selectQuery := fmt.Sprintf("SELECT * FROM santri %s%s LIMIT $%d OFFSET $%d", where, orderClause, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriQueryFailed, err)
	}
	defer rows.Close()

	var items []*entity.Santri
	for rows.Next() {
		s, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, domainerr.Wrap(constant.CodeSantriQueryFailed, err)
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &repository.SantriListResult{
		Items: items,
		Meta: repository.SantriListMeta{
			CurrentPage: currentPage,
			PerPage:     int64(limit),
			Total:       total,
			TotalPages:  totalPages,
		},
	}, nil
}

func (r *PostgresSantriRepository) findOne(ctx context.Context, query string, args ...any) (*entity.Santri, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanRow(row)
}

func (r *PostgresSantriRepository) scanRow(scanner interface{ Scan(...any) error }) (*entity.Santri, error) {
	var s entity.Santri
	var nis *string
	err := scanner.Scan(
		&s.ID, &s.UserID,
		&s.Nickname, &s.Program, &s.Option, &s.Hobby, &s.Purpose, &s.MotivationEntry,
		&s.POB, &s.DOB, &s.Blood,
		&s.Address, &s.SubDistrict, &s.District, &s.Province, &s.PostalCode,
		&s.PreviousPondokName, &s.PreviousPondokAddress, &s.PreviousPondokDiv, &s.PreviousPondokTime,
		&s.NIK, &s.NoKK, &s.NISN, &s.NoKIP, &s.NoKKS, &s.NoPKH,
		&s.Workplace, &s.Department,
		&s.HomeStatus,
		&s.Father, &s.FatherPN, &s.FatherNIK, &s.FatherJob, &s.FatherGraduate, &s.FatherIncome,
		&s.Mother, &s.MotherPN, &s.MotherNIK, &s.MotherJob, &s.MotherGraduate, &s.MotherIncome,
		&s.GuardianRelationship, &s.Guardian, &s.GuardianPN, &s.GuardianNIK, &s.GuardianJob, &s.GuardianGraduate, &s.GuardianIncome,
		&nis,
		&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New(constant.CodeSantriNotFound)
		}
		return nil, domainerr.Wrap(constant.CodeSantriQueryFailed, err)
	}
	if nis != nil && *nis != "" {
		n, voErr := santriVO.NewNIS(*nis)
		if voErr == nil {
			s.NIS = &n
		}
	}
	return &s, nil
}
