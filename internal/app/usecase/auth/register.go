package authUsecase

import (
	"context"
	"errors"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/dto"
	"sipon-api/internal/app/port"
	domainerr "sipon-api/internal/domain/errors"
	roleconstant "sipon-api/internal/domain/role/constant"
	roleentity "sipon-api/internal/domain/role/entity"
	rolerepo "sipon-api/internal/domain/role/repository"
	"sipon-api/internal/domain/user/constant"
	usererr "sipon-api/internal/domain/user/constant"
	"sipon-api/internal/domain/user/entity"
	"sipon-api/internal/domain/user/repository"
	"sipon-api/internal/domain/user/valueobject"
	verificationrepo "sipon-api/internal/domain/verification/repository"
	"time"

	"github.com/google/uuid"
)

type RegisterUseCase struct {
	userRepo     repository.UserRepository
	verifRepo    verificationrepo.VerificationRepository
	hasher       port.PasswordHasher
	otpGen       port.OTPGenerator
	emailSender  port.EmailSender
	smsSender    port.SMSSender
	tokenGen     port.TokenGenerator
	transactor   port.Transactor
	roleRepo     rolerepo.RoleRepository
	userRoleRepo rolerepo.UserRoleRepository
}

func NewRegisterUseCase(
	userRepo repository.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	hasher port.PasswordHasher,
	otpGen port.OTPGenerator,
	emailSender port.EmailSender,
	smsSender port.SMSSender,
	tokenGen port.TokenGenerator,
	transactor port.Transactor,
	roleRepo rolerepo.RoleRepository,
	userRoleRepo rolerepo.UserRoleRepository,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:     userRepo,
		verifRepo:    verifRepo,
		hasher:       hasher,
		otpGen:       otpGen,
		emailSender:  emailSender,
		smsSender:    smsSender,
		tokenGen:     tokenGen,
		transactor:   transactor,
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

// Required — role: public | perm: - | benefit: -
func (uc *RegisterUseCase) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 1. Validasi format via value object
	email, err := valueobject.NewEmail(req.Email)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodeInvalidEmailFormat:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	username, err := valueobject.NewUsername(req.Username)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodeInvalidUsernameFormat:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	plainPw, err := valueobject.NewPlainPassword(req.Password)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodePasswordTooShort, usererr.CodePasswordMustHaveUppercase, usererr.CodePasswordMustHaveDigit:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	var phone *valueobject.PhoneNumber
	if req.Phone != nil {
		phone, err = valueobject.NewPhoneNumber(*req.Phone)
		if err != nil {
			var de *domainerr.DomainError
			if errors.As(err, &de) {
				switch de.Code {
				case usererr.CodeInvalidPhoneNumberFormat:
					return nil, apperror.Unprocessable(string(de.Code), nil, err)
				}
			}
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
	}

	// 2. Cek duplikat
	exists, err := uc.userRepo.ExistsByLoginIdentity(ctx, constant.LoginIdentifierEmail, email.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	if exists {
		return nil, apperror.Conflict(string(apperror.CodeConflict))
	}
	if phone != nil {
		exists, err = uc.userRepo.ExistsByLoginIdentity(ctx, constant.LoginIdentifierPhone, phone.Value())
		if err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
		if exists {
			return nil, apperror.Conflict(string(apperror.CodeConflict))
		}
	}
	exists, err = uc.userRepo.ExistsByLoginIdentity(ctx, constant.LoginIdentifierUsername, username.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	if exists {
		return nil, apperror.Conflict(string(apperror.CodeConflict))
	}

	// 3. Hash password (infra concern, bukan domain)
	hashedStr, err := uc.hasher.Hash(plainPw.Value())
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	hashed, err := valueobject.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 4. Buat User aggregate
	userID := uuid.NewString()
	user, err := entity.NewUser(userID, username, req.Fullname, email, phone)
	if err != nil {
		var de *domainerr.DomainError
		if errors.As(err, &de) {
			switch de.Code {
			case usererr.CodeUserIDRequired, usererr.CodeUserEmailRequired, usererr.CodeUserPhoneNumberInvalid:
				return nil, apperror.Unprocessable(string(de.Code), nil, err)
			}
		}
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	now := time.Now()

	// 5. Buat ONE Credential (local provider) lalu attach semua login identities ke dalamnya.
	credID := uuid.NewString()
	cred := entity.NewLocalCredential(credID, userID, hashed, true)

	emailIdentity, err := entity.NewLoginIdentity(uuid.NewString(), userID, credID, constant.LoginIdentifierEmail, email.Value(), true, nil)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	cred.AddLoginIdentity(emailIdentity)

	usernameIdentity, err := entity.NewLoginIdentity(uuid.NewString(), userID, credID, constant.LoginIdentifierUsername, username.Value(), true, &now)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}
	cred.AddLoginIdentity(usernameIdentity)

	if phone != nil {
		phoneIdentity, err := entity.NewLoginIdentity(uuid.NewString(), userID, credID, constant.LoginIdentifierPhone, phone.Value(), true, nil)
		if err != nil {
			return nil, apperror.Internal(string(apperror.CodeInternal), err)
		}
		cred.AddLoginIdentity(phoneIdentity)
	}

	user.AddCredential(cred)

	// 6. Simpan user + assign role member (global) dalam satu transaksi
	memberRole, err := uc.roleRepo.FindByName(ctx, roleconstant.MemberRoleName)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	userRole, err := roleentity.NewUserRole(uuid.NewString(), userID, memberRole.ID, roleconstant.ScopeTypeGlobal, nil, userID, nil)
	if err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Save(txCtx, user); err != nil {
			return err
		}
		return uc.userRoleRepo.Save(txCtx, userRole)
	}); err != nil {
		return nil, apperror.Internal(string(apperror.CodeInternal), err)
	}

	// 7. Terbitkan token pair sekalian, supaya client bisa langsung login tanpa
	// panggilan /login terpisah.
	loginResp, err := issueTokenPair(user, req.DeviceID, uc.tokenGen)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		UserID:        userID,
		LoginResponse: *loginResp,
	}, nil
}
