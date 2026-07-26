package repository

import (
	"context"
	verificationConstant "sipon-api/internal/domain/verification/constant"
	verificationentity "sipon-api/internal/domain/verification/entity"
)

type VerificationRepository interface {
	Save(ctx context.Context, code *verificationentity.VerificationCode) error
	FindLatestByUserAndPurpose(ctx context.Context, userID string, purpose verificationConstant.CodePurpose) (*verificationentity.VerificationCode, error)
	Update(ctx context.Context, code *verificationentity.VerificationCode) error
}
