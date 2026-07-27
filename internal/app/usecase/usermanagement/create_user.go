package usermanagement

import (
	"context"
	"strings"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	userentity "sipon-api/internal/domain/user/entity"
	userrepo "sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"

	"github.com/google/uuid"
)

type CreateUserUseCase struct {
	userRepo userrepo.UserRepository
	hasher   port.PasswordHasher
}

func NewCreateUserUseCase(userRepo userrepo.UserRepository, hasher port.PasswordHasher) *CreateUserUseCase {
	return &CreateUserUseCase{userRepo: userRepo, hasher: hasher}
}

// Required — perm: manage_users
//
// Mirip RegisterUseCase tetapi: tanpa verifikasi OTP, tanpa auto-assign role
// member (admin assign roles afterward via user-roles endpoint). Password
// auto-generated, dikembalikan satu kali di response (lihat plan §Decision 1).
func (uc *CreateUserUseCase) Execute(ctx context.Context, req dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	// 1. Validasi format via value object.
	email, err := valueobject.NewEmail(req.Email)
	if err != nil {
		return nil, mapUserValidationError(err)
	}
	username, err := valueobject.NewUsername(req.Username)
	if err != nil {
		return nil, mapUserValidationError(err)
	}
	var phone *valueobject.PhoneNumber
	if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		phone, err = valueobject.NewPhoneNumber(*req.Phone)
		if err != nil {
			return nil, mapUserValidationError(err)
		}
	}

	// 2. Generate password temporer yang memenuhi aturan PlainPassword.
	plainStr, err := generateTemporaryPassword()
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	plain, err := valueobject.NewPlainPassword(plainStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 3. Hash password (infra concern, bukan domain).
	hashedStr, err := uc.hasher.Hash(plain.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	hashed, err := valueobject.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 4. Bangun User aggregate.
	userID := uuid.NewString()
	user, err := userentity.NewUser(userID, username, req.Fullname, email, phone)
	if err != nil {
		return nil, mapUserValidationError(err)
	}

	// 5. Credential + login identities (verified=false; admin lewati verifikasi).
	cred, err := buildCredentialWithIdentities(
		userID, hashed, username.Value(), email.Value(),
		phoneValuePtr(phone),
	)
	if err != nil {
		return nil, mapUserValidationError(err)
	}
	user.AddCredential(cred)

	// 6. Save aggregate. Tidak terbungkus transaksi (sesuai plan §5 —
	// admin-create memakai Save langsung). Konstrain DB jadi backstop untuk
	// race TOCTOU existence check (lihat plan §4.1).
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, mapUserDomainError(err)
	}

	// 7. Build response + generated_password sekali.
	resp, err := buildUserManagementResponse(ctx, nil, user, false)
	if err != nil {
		return nil, err
	}
	return &dto.CreateUserResponse{
		UserManagementResponse: *resp,
		GeneratedPassword:      plain.Value(),
	}, nil
}

// phoneValuePtr mengembalikan *string value dari PhoneNumber untuk diteruskan
// ke buildCredentialWithIdentities.
func phoneValuePtr(p *valueobject.PhoneNumber) *string {
	if p == nil {
		return nil
	}
	v := p.Value()
	return &v
}