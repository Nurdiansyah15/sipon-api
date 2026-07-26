package persistence

import (
	"context"
	"database/sql"
	"time"

	domainerr "sipon-api/internal/domain/errors"
	"sipon-api/internal/domain/user/valueobject"
	verificationConstant "sipon-api/internal/domain/verification/constant"
	verificationentity "sipon-api/internal/domain/verification/entity"
)

// ── Struct & Constructor ──────────────────────────────────────────────────────

type PostgresVerificationRepository struct {
	db *sql.DB
}

func NewPostgresVerificationRepository(db *sql.DB) *PostgresVerificationRepository {
	return &PostgresVerificationRepository{db: db}
}

// ── Methods ───────────────────────────────────────────────────────────────────

func (r *PostgresVerificationRepository) Save(ctx context.Context, code *verificationentity.VerificationCode) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx, `
		INSERT INTO verification_codes (id, user_id, code, new_identity_value, purpose, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		code.ID, code.UserID, code.Code.Value(), code.NewIdentityValue, string(code.Purpose), code.ExpiresAt, code.CreatedAt,
	)
	if err != nil {
		return domainerr.Wrap(verificationConstant.CodeVerificationPersistenceFailed, err)
	}
	return nil
}

func (r *PostgresVerificationRepository) FindLatestByUserAndPurpose(ctx context.Context, userID string, purpose verificationConstant.CodePurpose) (*verificationentity.VerificationCode, error) {
	row := execFromContext(ctx, r.db).QueryRowContext(ctx, `
		SELECT id, user_id, code, new_identity_value, purpose, expires_at, used_at, created_at
		FROM verification_codes
		WHERE user_id=$1 AND purpose=$2
		ORDER BY created_at DESC LIMIT 1`,
		userID, string(purpose),
	)
	var (
		id, uid, code, purp  string
		newIdentityValue     sql.NullString
		expiresAt, createdAt time.Time
		usedAt               sql.NullTime
	)
	if err := row.Scan(&id, &uid, &code, &newIdentityValue, &purp, &expiresAt, &usedAt, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerr.New(verificationConstant.CodeVerificationNotFound)
		}
		return nil, domainerr.Wrap(verificationConstant.CodeVerificationQueryFailed, err)
	}
	otpCode, err := valueobject.NewOTPCode(code)
	if err != nil {
		return nil, domainerr.Wrap(verificationConstant.CodeOTPInvalid, err)
	}
	vc := &verificationentity.VerificationCode{
		ID:        id,
		UserID:    uid,
		Code:      otpCode,
		Purpose:   verificationConstant.CodePurpose(purp),
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}
	if newIdentityValue.Valid {
		vc.NewIdentityValue = &newIdentityValue.String
	}
	if usedAt.Valid {
		vc.UsedAt = &usedAt.Time
	}
	return vc, nil
}

func (r *PostgresVerificationRepository) Update(ctx context.Context, code *verificationentity.VerificationCode) error {
	_, err := execFromContext(ctx, r.db).ExecContext(ctx,
		`UPDATE verification_codes SET used_at=$1 WHERE id=$2`,
		code.UsedAt, code.ID,
	)
	if err != nil {
		return domainerr.Wrap(verificationConstant.CodeVerificationPersistenceFailed, err)
	}
	return nil
}
