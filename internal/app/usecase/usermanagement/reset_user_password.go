package usermanagement

import (
	"context"
	"strings"
	"time"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	userconstant "sipon-api/internal/domain/user/constant"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
)

type ResetUserPasswordUseCase struct {
	userRepo userrepo.UserRepository
	hasher   port.PasswordHasher
}

func NewResetUserPasswordUseCase(userRepo userrepo.UserRepository, hasher port.PasswordHasher) *ResetUserPasswordUseCase {
	return &ResetUserPasswordUseCase{userRepo: userRepo, hasher: hasher}
}

// Required — perm: reset_user_password
//
// Generate+hash password temporer baru, reset failed-login attempts. Password
// dikembalikan satu kali di response.
func (uc *ResetUserPasswordUseCase) Execute(ctx context.Context, userID string) (*dto.ResetUserPasswordResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return nil, mapUserDomainError(err)
	}

	plainStr, err := generateTemporaryPassword()
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	plain, err := valueobject.NewPlainPassword(plainStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	hashedStr, err := uc.hasher.Hash(plain.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	newHashed, err := valueobject.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// Mutate SecretHash pada credential LOCAL yang aktif (sama dengan alur
	// reset_password self-service — lihat auth/reset_password.go).
	local := user.FindCredential(userconstant.CredentialTypeLocal)
	if local == nil || local.DeletedAt != nil {
		return nil, apperror.Forbidden(string(userconstant.CodeUserNoLocalCredential))
	}
	now := time.Now()
	local.SecretHash = &newHashed
	local.LastChangedAt = &now
	local.UpdatedAt = now
	user.UpdatedAt = now
	user.ResetFailedAttempts()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, mapUserDomainError(err)
	}

	return &dto.ResetUserPasswordResponse{GeneratedPassword: plain.Value()}, nil
}